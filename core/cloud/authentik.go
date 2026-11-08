package cloud

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/madhank93/homelab/config"
	"github.com/madhank93/homelab/core/internal/cfg"
	"github.com/pulumi/pulumi-terraform-provider/sdks/go/authentik/v2026/authentik"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// AuthURL is the public Authentik base URL on Bifrost.
var AuthURL = "https://" + config.C.Host("auth")

// GithubName, GithubSlug and GithubClientID identify the GitHub OAuth app used
// as the login source; its secret comes from AUTHENTIK_GITHUB_SECRET.
const (
	GithubName     = "GitHub"
	GithubSlug     = "github"
	GithubClientID = "Ov23liUPVh4nPuUJzGFp"
)

// identStageInfo is the subset of an identification stage returned by the
// Authentik REST API.
type identStageInfo struct {
	PK         string   `json:"pk"`
	UserFields []string `json:"user_fields"`
}

// getDefaultIdentificationStage fetches the default-authentication-identification
// stage from the Authentik REST API. The SDK has no Lookup for stages, so its
// UUID is needed to import the stage into Pulumi state.
func getDefaultIdentificationStage(apiURL, token string) (*identStageInfo, error) {
	url := apiURL + "/api/v3/stages/identification/?name=default-authentication-identification"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET identification stage: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read identification stage response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET identification stage HTTP %d: %s", resp.StatusCode, body)
	}

	var result struct {
		Results []identStageInfo `json:"results"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse identification stage response: %w", err)
	}
	if len(result.Results) == 0 {
		return nil, fmt.Errorf("default-authentication-identification stage not found in Authentik")
	}
	return &result.Results[0], nil
}

// AuthentikContext carries shared Authentik provider state and pre-resolved
// flow/scope references that are reused across all application registrations.
type AuthentikContext struct {
	Ctx            *pulumi.Context
	Provider       *authentik.Provider
	FlowAuth       pulumi.StringInput // default-authentication-flow
	FlowInvalid    pulumi.StringInput // default-provider-invalidation-flow
	FlowImplicit   pulumi.StringInput // implicit-consent authorization flow
	FlowExplicit   pulumi.StringInput // explicit-consent authorization flow
	FlowEnrollment pulumi.StringInput // default-source-enrollment (required for OAuth sources)
	Scopes         pulumi.StringArray
	SigningKey     pulumi.StringInput // authentik Self-signed Certificate ID
}

// OIDCApp describes an OIDC provider and application to register in Authentik.
type OIDCApp struct {
	Name                  string
	Slug                  string
	ClientID              string
	ClientSecret          string // required when ClientType is "confidential"
	ClientType            string // "public" or "confidential"; defaults to "public"
	Redirects             []string
	LaunchURL             string
	ExtraPropertyMappings pulumi.StringArray // additional scope/property mappings beyond the defaults
	AccessGroup           pulumi.StringInput // when set, only this group's members may log in
	GrantTypes            []string           // defaults to authorization_code + refresh_token
}

// DeployAuthentik configures the Authentik identity provider running on Bifrost.
// It registers GitHub as a login source, creates the OIDC applications (Grafana,
// Homepage, Argo CD, Harbor, Kubernetes/Headlamp, Proxmox, NetBird) and the
// ForwardAuth proxy provider that Traefik uses for public services.
//
// AUTHENTIK_TOKEN, AUTHENTIK_GITHUB_SECRET and each app's client secret must be in
// the SOPS-injected environment.
func DeployAuthentik(ctx *pulumi.Context) error {
	token := cfg.K.String("AUTHENTIK_TOKEN")
	ghSecret := cfg.K.String("AUTHENTIK_GITHUB_SECRET")
	if token == "" || ghSecret == "" {
		return fmt.Errorf("missing AUTHENTIK_TOKEN or AUTHENTIK_GITHUB_SECRET")
	}

	provider, err := authentik.NewProvider(ctx, "authentik-provider", &authentik.ProviderArgs{
		Url:   pulumi.String(AuthURL),
		Token: pulumi.String(token),
		// Explicit so a stray AUTHENTIK_INSECURE env var cannot turn off TLS verification.
		Insecure: pulumi.Bool(false),
	})
	if err != nil {
		return err
	}

	ac := AuthentikContext{Ctx: ctx, Provider: provider}
	for _, f := range []struct {
		dst  *pulumi.StringInput
		slug string
	}{
		{&ac.FlowAuth, "default-authentication-flow"},
		{&ac.FlowInvalid, "default-provider-invalidation-flow"},
		{&ac.FlowImplicit, "default-provider-authorization-implicit-consent"},
		{&ac.FlowExplicit, "default-provider-authorization-explicit-consent"},
		{&ac.FlowEnrollment, "default-source-enrollment"},
	} {
		if *f.dst, err = lookupFlow(ctx, provider, f.slug); err != nil {
			return err
		}
	}

	// Authentik creates this key pair on install; OIDC clients such as NetBird
	// verify ID tokens against it.
	signingKey, err := authentik.LookupCertificateKeyPair(ctx, &authentik.LookupCertificateKeyPairArgs{
		Name: "authentik Self-signed Certificate",
	}, pulumi.Provider(provider))
	if err != nil {
		return fmt.Errorf("failed to find default signing key: %w", err)
	}

	ac.SigningKey = pulumi.String(signingKey.Id)

	// NetBird requests an "api" scope that Authentik does not ship; it carries no claims.
	// See https://integrations.goauthentik.io/networking/netbird/.
	apiScope, err := authentik.NewPropertyMappingProviderScope(ctx, "scope-api", &authentik.PropertyMappingProviderScopeArgs{
		Name:       pulumi.String("Netbird API Scope"),
		ScopeName:  pulumi.String("api"),
		Expression: pulumi.String("return {}"),
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	allScopes, err := getCommonScopes(ctx, provider)
	if err != nil {
		return err
	}
	allScopes = append(allScopes,
		apiScope.ID().ToStringOutput(),
	)
	ac.Scopes = allScopes

	githubSource, err := authentik.NewSourceOauth(ctx, "source-github", &authentik.SourceOauthArgs{
		Name:               pulumi.String(GithubName),
		Slug:               pulumi.String(GithubSlug),
		AuthenticationFlow: ac.FlowAuth,
		EnrollmentFlow:     ac.FlowEnrollment,
		ProviderType:       pulumi.String("github"),
		ConsumerKey:        pulumi.String(GithubClientID),
		ConsumerSecret:     pulumi.String(ghSecret),
		Pkce:               pulumi.String("S256"),
		// "identifier" creates a new Authentik user from the GitHub identity without
		// trying to link to an existing user by email. "email_link" requires the user
		// to already be authenticated in Authentik to confirm the link — which loops
		// when the admin email matches the GitHub email but the user isn't logged in.
		UserMatchingMode: pulumi.String("identifier"),
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	// Binding the GitHub source to the default identification stage shows the
	// "Login with GitHub" button. The stage is adopted with pulumi.Import, a no-op
	// once it is in state; its user_fields are carried over unchanged.
	identStage, err := getDefaultIdentificationStage(AuthURL, token)
	if err != nil {
		return fmt.Errorf("fetch default identification stage: %w", err)
	}
	userFieldInputs := make(pulumi.StringArray, len(identStage.UserFields))
	for i, f := range identStage.UserFields {
		userFieldInputs[i] = pulumi.String(f)
	}

	_, err = authentik.NewStageIdentification(ctx, "stage-default-identification",
		&authentik.StageIdentificationArgs{
			Name:       pulumi.String("default-authentication-identification"),
			UserFields: userFieldInputs,
			Sources:    pulumi.StringArray{githubSource.Uuid},
		},
		pulumi.Provider(provider),
		pulumi.Import(pulumi.ID(identStage.PK)),
		pulumi.DependsOn([]pulumi.Resource{githubSource}),
	)
	if err != nil {
		return err
	}

	// --- Applications ---

	// The groups claim lists the user's group names; apps map grafana-admins to
	// their admin role.
	groupsMapping, err := authentik.NewPropertyMappingProviderScope(ctx, "scope-groups", &authentik.PropertyMappingProviderScopeArgs{
		Name:      pulumi.String("Groups"),
		ScopeName: pulumi.String("groups"),
		Expression: pulumi.String(
			`return [group.name for group in request.user.ak_groups.all()]`,
		),
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	// grafana-admins group — members get Admin in Grafana's main org via org_mapping.
	grafanaAdmins, err := lookupGrafanaAdmins(ctx, provider)
	if err != nil {
		return err
	}
	grafanaAdminsGroup, err := authentik.NewGroup(ctx, "group-grafana-admins", &authentik.GroupArgs{
		Name:  pulumi.String("grafana-admins"),
		Users: grafanaAdmins,
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	// Grafana reads its client secret from OpenBao (secret/grafana, seeded from
	// SOPS by just openbao-setup).
	grafanaSecret := cfg.K.String("GRAFANA_OAUTH_CLIENT_SECRET")
	if grafanaSecret == "" {
		return fmt.Errorf("missing GRAFANA_OAUTH_CLIENT_SECRET in config")
	}
	if err := createOIDCApp(ac, OIDCApp{
		Name:         "Grafana",
		Slug:         "grafana",
		ClientID:     "grafana-homelab",
		ClientSecret: grafanaSecret,
		ClientType:   "confidential",
		LaunchURL:    "https://" + config.C.Host("grafana") + "/",
		Redirects:    []string{"https://" + config.C.Host("grafana") + "/login/generic_oauth"},
		ExtraPropertyMappings: pulumi.StringArray{
			groupsMapping.ID().ToStringOutput(),
		},
	}); err != nil {
		return err
	}
	ctx.Export("GrafanaOIDCClientID", pulumi.String("grafana-homelab"))

	// Homepage admits any identity the provider authorizes, and any GitHub account can enroll in Authentik,
	// so access is limited to the grafana-admins group.
	homepageSecret := cfg.K.String("HOMEPAGE_OAUTH_CLIENT_SECRET")
	if homepageSecret == "" {
		return fmt.Errorf("missing HOMEPAGE_OAUTH_CLIENT_SECRET in config")
	}
	if err := createOIDCApp(ac, OIDCApp{
		Name:         "Homepage",
		Slug:         "homepage",
		ClientID:     "homepage-homelab",
		ClientSecret: homepageSecret,
		ClientType:   "confidential",
		LaunchURL:    "https://" + config.C.Host("home") + "/",
		Redirects:    []string{"https://" + config.C.Host("home") + "/api/auth/callback/homepage-oidc"},
		AccessGroup:  grafanaAdminsGroup.ID().ToStringOutput(),
	}); err != nil {
		return err
	}

	// Admin UIs that map the groups claim to their own admin role. Each client
	// secret is in SOPS as <SECRET>; the app reads it from OpenBao (Harbor,
	// Headlamp) or a Pulumi-made Secret (Argo CD, which deploys OpenBao).
	// "kubernetes" is the kube-apiserver's OIDC audience; Headlamp logs in with it.
	// Proxmox redirects back to its own origin, port included.
	pve := "https://" + config.C.Host(config.C.Proxmox.Host) + ":8006"
	for _, a := range []struct{ name, slug, clientID, secret, base, callback string }{
		{"Argo CD", "argocd", "argocd-homelab", "ARGOCD_OAUTH_CLIENT_SECRET", "https://" + config.C.Host("argocd"), "/auth/callback"},
		{"Harbor", "harbor", "harbor-homelab", "HARBOR_OAUTH_CLIENT_SECRET", "https://" + config.C.Host("harbor"), "/c/oidc/callback"},
		{"Kubernetes", "kubernetes", "kubernetes", "KUBERNETES_OAUTH_CLIENT_SECRET", "https://" + config.C.Host("headlamp"), "/oidc-callback"},
		{"Proxmox", "proxmox", "proxmox-homelab", "PROXMOX_OAUTH_CLIENT_SECRET", pve, ""},
	} {
		secret := cfg.K.String(a.secret)
		if secret == "" {
			return fmt.Errorf("missing %s in config", a.secret)
		}
		base := a.base
		if err := createOIDCApp(ac, OIDCApp{
			Name:                  a.name,
			Slug:                  a.slug,
			ClientID:              a.clientID,
			ClientSecret:          secret,
			ClientType:            "confidential",
			LaunchURL:             base + "/",
			Redirects:             []string{base + a.callback},
			ExtraPropertyMappings: pulumi.StringArray{groupsMapping.ID().ToStringOutput()},
			AccessGroup:           grafanaAdminsGroup.ID().ToStringOutput(),
		}); err != nil {
			return err
		}
	}

	// NetBird's embedded Dex federates to Authentik; the Dex connector itself is
	// registered by hand in the NetBird UI.
	netbirdSecret := cfg.K.String("NETBIRD_CLIENT_SECRET")
	if netbirdSecret == "" {
		return fmt.Errorf("missing NETBIRD_CLIENT_SECRET in config")
	}
	if err := createOIDCApp(ac, OIDCApp{
		Name:         "Netbird",
		Slug:         "netbird",
		ClientID:     "aumenijDycfG1cQURqH9BNJpV3KVUCoMHGPUVUlT",
		ClientSecret: netbirdSecret,
		ClientType:   "confidential",
		LaunchURL:    "https://" + config.C.Host("netbird") + "/",
		Redirects: []string{
			// Dex callback: its issuer (/oauth2) + "/callback".
			"https://" + config.C.Host("netbird") + "/oauth2/callback",
			// NetBird CLI login callback.
			"http://localhost:53000",
		},
		GrantTypes: []string{"authorization_code", "refresh_token", "urn:ietf:params:oauth:grant-type:device_code"},
	}); err != nil {
		return err
	}

	if err := createHomelabForwardAuth(ac); err != nil {
		return err
	}

	return nil
}

// lookupGrafanaAdmins resolves config.C.GrafanaAdmins emails to Authentik user
// PKs, skipping (with a warning) users who have not logged in yet.
func lookupGrafanaAdmins(ctx *pulumi.Context, provider *authentik.Provider) (pulumi.Float64Array, error) {
	var grafanaAdmins pulumi.Float64Array
	for _, email := range config.C.GrafanaAdmins {
		res, err := authentik.GetUsers(ctx, &authentik.GetUsersArgs{Email: pulumi.StringRef(email)}, pulumi.Provider(provider))
		if err != nil {
			return nil, err
		}
		if len(res.Users) == 0 {
			_ = ctx.Log.Warn("grafana_admins: no Authentik user "+email+" yet; log in once, then re-run", nil)
			continue
		}
		grafanaAdmins = append(grafanaAdmins, pulumi.Float64(res.Users[0].Pk))
	}
	return grafanaAdmins, nil
}

// createHomelabForwardAuth creates the Authentik forward-domain proxy provider and
// outpost that back Traefik's ForwardAuth middleware. Every public service
// passes through it unless its publicServices entry sets SkipAuth.
func createHomelabForwardAuth(ac AuthentikContext) error {
	proxyProvider, err := authentik.NewProviderProxy(ac.Ctx, "provider-homelab-forwardauth", &authentik.ProviderProxyArgs{
		Name:              pulumi.String("Homelab ForwardAuth"),
		Mode:              pulumi.String("forward_domain"),
		ExternalHost:      pulumi.String("https://" + config.C.Domain),
		CookieDomain:      pulumi.String("." + config.C.Domain),
		AuthorizationFlow: ac.FlowImplicit,
		InvalidationFlow:  ac.FlowInvalid,
	}, pulumi.Provider(ac.Provider))
	if err != nil {
		return err
	}

	_, err = authentik.NewApplication(ac.Ctx, "app-homelab-forwardauth", &authentik.ApplicationArgs{
		Name:             pulumi.String("Homelab ForwardAuth"),
		Slug:             pulumi.String("homelab-forwardauth"),
		ProtocolProvider: idFloat(proxyProvider.ID()).ToFloat64PtrOutput(),
	}, pulumi.Provider(ac.Provider))
	if err != nil {
		return err
	}

	proxyProviderID := idFloat(proxyProvider.ID())

	_, err = authentik.NewOutpost(ac.Ctx, "outpost-homelab-embedded", &authentik.OutpostArgs{
		Name: pulumi.String("Homelab Embedded Outpost"),
		Type: pulumi.String("proxy"),
		ProtocolProviders: pulumi.Float64Array{
			proxyProviderID,
		},
	}, pulumi.Provider(ac.Provider))
	return err
}

// createOIDCApp registers an OIDC provider and its application in Authentik,
// with ac.Scopes plus app.ExtraPropertyMappings. Redirect URIs containing "*"
// match as regex. When app.AccessGroup is set, a policy binding restricts login
// to that group.
func createOIDCApp(ac AuthentikContext, app OIDCApp) error {
	var redirectMap pulumi.StringMapArray
	for _, url := range app.Redirects {
		mode := "strict"
		if strings.Contains(url, "*") {
			mode = "regex"
		}
		redirectMap = append(redirectMap, pulumi.StringMap{
			"matching_mode": pulumi.String(mode),
			"url":           pulumi.String(url),
		})
	}

	// Authentik 2026.5+ rejects any grant type not listed, and lists none by default.
	grantTypes := app.GrantTypes
	if len(grantTypes) == 0 {
		grantTypes = []string{"authorization_code", "refresh_token"}
	}

	clientType := app.ClientType
	if clientType == "" {
		clientType = "public"
	}

	// slices.Concat allocates, so apps never share ac.Scopes' backing array.
	scopes := slices.Concat(ac.Scopes, app.ExtraPropertyMappings)

	providerArgs := &authentik.ProviderOauth2Args{
		Name:                   pulumi.String(app.Name),
		ClientId:               pulumi.String(app.ClientID),
		ClientType:             pulumi.String(clientType),
		AuthorizationFlow:      ac.FlowExplicit,
		InvalidationFlow:       ac.FlowInvalid,
		PropertyMappings:       scopes,
		AllowedRedirectUris:    redirectMap,
		SubMode:                pulumi.String("user_id"),
		IncludeClaimsInIdToken: pulumi.Bool(true),
		GrantTypes:             pulumi.ToStringArray(grantTypes),
		// Without a key pair, ID tokens are HS256-signed with the client secret,
		// which NetBird cannot verify.
		SigningKey: ac.SigningKey,
		// Lifetimes follow NetBird's Authentik guide.
		AccessCodeValidity:  pulumi.String("minutes=10"),
		AccessTokenValidity: pulumi.String("hours=1"),
	}

	if app.ClientSecret != "" {
		providerArgs.ClientSecret = pulumi.String(app.ClientSecret)
	}

	oidcProvider, err := authentik.NewProviderOauth2(ac.Ctx, fmt.Sprintf("provider-%s", app.ClientID), providerArgs, pulumi.Provider(ac.Provider))
	if err != nil {
		return err
	}

	providerIDPtr := idFloat(oidcProvider.ID()).ToFloat64PtrOutput()

	application, err := authentik.NewApplication(ac.Ctx, fmt.Sprintf("app-%s", app.ClientID), &authentik.ApplicationArgs{
		Name:             pulumi.String(app.Name),
		Slug:             pulumi.String(app.Slug),
		ProtocolProvider: providerIDPtr,
		MetaLaunchUrl:    pulumi.String(app.LaunchURL),
	}, pulumi.Provider(ac.Provider))
	if err != nil || app.AccessGroup == nil {
		return err
	}

	_, err = authentik.NewPolicyBinding(ac.Ctx, fmt.Sprintf("binding-%s", app.ClientID), &authentik.PolicyBindingArgs{
		Target: application.Uuid,
		Group:  app.AccessGroup,
		Order:  pulumi.Float64(0),
	}, pulumi.Provider(ac.Provider))
	return err
}

// lookupFlow looks up an Authentik flow by slug. Authentik ships these default
// flows on every install; their absence means no authentication can work.
func lookupFlow(ctx *pulumi.Context, prov *authentik.Provider, slug string) (pulumi.StringInput, error) {
	flow, err := authentik.LookupFlow(ctx, &authentik.LookupFlowArgs{
		Slug: pulumi.StringRef(slug),
	}, pulumi.Provider(prov))
	if err != nil {
		return nil, fmt.Errorf("look up authentik flow %q: %w", slug, err)
	}
	return pulumi.String(flow.Id), nil
}

// getCommonScopes returns the Authentik property-mapping IDs for the standard
// OIDC scopes (email, profile, openid, offline_access).
func getCommonScopes(ctx *pulumi.Context, prov *authentik.Provider) (pulumi.StringArray, error) {
	var scopeIDs pulumi.StringArray

	names := []string{"email", "profile", "openid", "offline_access"}

	for _, name := range names {
		scope, err := authentik.LookupPropertyMappingProviderScope(ctx, &authentik.LookupPropertyMappingProviderScopeArgs{
			Managed: pulumi.StringRef(fmt.Sprintf("goauthentik.io/providers/oauth2/scope-%s", name)),
		}, pulumi.Provider(prov))
		if err != nil {
			return nil, fmt.Errorf("look up authentik scope %q: %w", name, err)
		}
		scopeIDs = append(scopeIDs, pulumi.String(scope.Id))
	}

	return scopeIDs, nil
}

// idFloat converts a Pulumi resource ID to the numeric form the Authentik
// provider takes for cross-resource references.
func idFloat(id pulumi.IDOutput) pulumi.Float64Output {
	return id.ApplyT(func(id pulumi.ID) (float64, error) {
		n, err := strconv.Atoi(string(id))
		if err != nil {
			return 0, fmt.Errorf("non-numeric authentik ID %q: %w", id, err)
		}
		return float64(n), nil
	}).(pulumi.Float64Output)
}
