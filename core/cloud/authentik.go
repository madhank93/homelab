package cloud

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/madhank93/homelab/core/internal/cfg"
	"github.com/pulumi/pulumi-terraform-provider/sdks/go/authentik/v2025/authentik"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	// Authentik System
	AuthUrl      = "https://auth.madhan.app"
	AuthInsecure = false

	// GitHub SSO Settings
	GithubName     = "GitHub"
	GithubSlug     = "github"
	GithubClientId = "Ov23liUPVh4nPuUJzGFp"
)

// identStageInfo holds the fields we need from the Authentik REST API response
// for the default-authentication-identification stage.
type identStageInfo struct {
	PK         string   `json:"pk"`
	UserFields []string `json:"user_fields"`
}

// getDefaultIdentificationStage fetches the default-authentication-identification
// stage from the Authentik REST API and returns its UUID and user_fields.
// Used to import and update the stage via Pulumi without a Lookup data source.
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
		return nil, err
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

type AuthentikContext struct {
	Ctx          *pulumi.Context
	Provider     *authentik.Provider
	FlowAuth     pulumi.StringInput
	FlowInvalid  pulumi.StringInput
	FlowImplicit pulumi.StringInput
	FlowExplicit pulumi.StringInput
	Scopes       pulumi.StringArrayInput
	SigningKey   pulumi.StringInput
}

type OIDCApp struct {
	Name         string
	Slug         string
	ClientID     string
	ClientSecret string // Optional: Required if ClientType is confidential
	ClientType   string // "public" (default) or "confidential"
	Redirects    []string
	LaunchURL    string
}

func DeployAuthentik(ctx *pulumi.Context) error {
	// Fetch Secrets
	token := cfg.K.String("AUTHENTIK_TOKEN")
	ghSecret := cfg.K.String("AUTHENTIK_GITHUB_SECRET")
	// CraneSecret removed as it was only for Arcane
	if token == "" || ghSecret == "" {
		return fmt.Errorf("missing AUTHENTIK_TOKEN or AUTHENTIK_GITHUB_SECRET")
	}

	// Safe Debug Logging
	fmt.Printf("DEBUG: Authentik URL: %s\n", AuthUrl)
	fmt.Printf("DEBUG: Authentik Token Present: %v, Length: %d\n", token != "", len(token))
	fmt.Printf("DEBUG: GitHub Secret Present: %v, Length: %d\n", ghSecret != "", len(ghSecret))
	fmt.Printf("DEBUG: GitHub Config -> Name: %s, Slug: %s, ClientID: %s\n", GithubName, GithubSlug, GithubClientId)

	// Setup Provider
	provider, err := authentik.NewProvider(ctx, "authentik-provider", &authentik.ProviderArgs{
		Url:      pulumi.String(AuthUrl),
		Token:    pulumi.String(token),
		Insecure: pulumi.Bool(AuthInsecure),
	})
	if err != nil {
		return err
	}

	// Global Flow Lookups
	ac := AuthentikContext{
		Ctx:          ctx,
		Provider:     provider,
		FlowAuth:     mustLookupFlow(ctx, provider, "default-authentication-flow"),
		FlowInvalid:  mustLookupFlow(ctx, provider, "default-provider-invalidation-flow"),
		FlowImplicit: mustLookupFlow(ctx, provider, "default-provider-authorization-implicit-consent"),
		FlowExplicit: mustLookupFlow(ctx, provider, "default-provider-authorization-explicit-consent"),
	}

	// Signing Key (Self-signed default)
	// We need a signing key for OIDC to work correctly with Netbird (JWTs).
	// Typically Authentik creates a default one. We'll search for it or use a known one.
	// For now, let's look up the default one created by Authentik.
	signingKey, err := authentik.LookupCertificateKeyPair(ctx, &authentik.LookupCertificateKeyPairArgs{
		Name: "authentik Self-signed Certificate",
	}, pulumi.Provider(provider))
	if err != nil {
		// Fallback or error if critical. It usually exists.
		// Use a panic here because if we can't sign tokens, OIDC won't work.
		return fmt.Errorf("failed to find default signing key: %w", err)
	}

	ac.SigningKey = pulumi.String(signingKey.Id)

	// Netbird requires "api" scope
	// Docs: https://integrations.goauthentik.io/networking/netbird/
	// We create a custom scope mapping for "api" which Netbird requests.
	apiScope, err := authentik.NewPropertyMappingProviderScope(ctx, "scope-api", &authentik.PropertyMappingProviderScopeArgs{
		Name:       pulumi.String("Netbird API Scope"),
		ScopeName:  pulumi.String("api"),
		Expression: pulumi.String("return {}"), // Standard empty mapping
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	// Common OIDC Scopes (email, profile, openid, offline_access)
	allScopes := getCommonScopes(ctx, provider)
	// Ensure we convert IDOutput to StringOutput for compatibility
	ac.Scopes = append(allScopes,
		apiScope.ID().ToStringOutput(),
	)

	// Add to Admins Group (needed for the service account)
	adminGroup, err := authentik.LookupGroup(ctx, &authentik.LookupGroupArgs{
		Name: pulumi.StringRef("authentik Admins"),
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	// Step 3 & 4: Service Account for Netbird (Required for sync)
	// We use built-in User resource with type=service_account
	sa, err := authentik.NewUser(ctx, "sa-netbird", &authentik.UserArgs{
		Username: pulumi.String("sa-netbird"),
		Name:     pulumi.String("Netbird"),
		Type:     pulumi.String("service_account"),
		Groups: pulumi.StringArray{
			pulumi.String(adminGroup.Id),
		},
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	// Generate App Token for Netbird
	// Note: We'll output this so the user can put it in management.json
	saToken, err := authentik.NewToken(ctx, "token-netbird", &authentik.TokenArgs{
		Identifier: pulumi.String("netbird-api-token"),
		User: sa.ID().ToStringOutput().ApplyT(func(id string) (float64, error) {
			idInt, err := strconv.Atoi(id)
			if err != nil {
				return 0.0, err
			}
			return float64(idInt), nil
		}).(pulumi.Float64Output), // Token User expects Float64Input
		Intent:   pulumi.String("app_password"),
		Expiring: pulumi.Bool(false),
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}
	ctx.Export("NetbirdServiceToken", saToken.Key)
	ctx.Export("NetbirdClientID", pulumi.String("aumenijDycfG1cQURqH9BNJpV3KVUCoMHGPUVUlT"))

	// GitHub SSO Source
	githubSource, err := authentik.NewSourceOauth(ctx, "source-github", &authentik.SourceOauthArgs{
		Name:               pulumi.String(GithubName),
		Slug:               pulumi.String(GithubSlug),
		AuthenticationFlow: ac.FlowAuth,
		// EnrollmentFlow:     enrollFlow, // Removed as requested
		ProviderType:     pulumi.String("github"),
		ConsumerKey:      pulumi.String(GithubClientId),
		ConsumerSecret:   pulumi.String(ghSecret),
		Pkce:             pulumi.String("S256"),
		UserMatchingMode: pulumi.String("email_link"),
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	// Bind GitHub source to the default identification stage so the "Login with
	// GitHub" button appears on the Authentik login page.
	// We fetch the stage UUID via REST API (no Lookup data source exists in the SDK),
	// then import + update the stage. pulumi.Import is a no-op on re-runs once the
	// resource is already in Pulumi state.
	identStage, err := getDefaultIdentificationStage(AuthUrl, token)
	if err != nil {
		return fmt.Errorf("fetch default identification stage: %w", err)
	}
	fmt.Printf("DEBUG: Identification stage UUID: %s, UserFields: %v\n", identStage.PK, identStage.UserFields)

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

	// Netbird — confidential OIDC app used by embedded Dex as an upstream connector.
	// The combined server always runs embedded Dex; users authenticate against Dex,
	// which federates to Authentik. Dex's callback URI is always issuer + "/callback".
	// Configure this connector: NetBird → Settings → Identity Providers → Add → Authentik
	//   Client ID:     aumenijDycfG1cQURqH9BNJpV3KVUCoMHGPUVUlT
	//   Client Secret: NETBIRD_CLIENT_SECRET from sops
	//   Issuer:        https://auth.madhan.app/application/o/netbird/
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
		LaunchURL:    "https://netbird.madhan.app/",
		Redirects: []string{
			// Embedded Dex callback: issuer (/oauth2) + /callback
			"https://netbird.madhan.app/oauth2/callback",
			// NetBird CLI device-auth callback
			"http://localhost:53000",
		},
	}); err != nil {
		return err
	}

	// Homelab ForwardAuth — covers *.madhan.app public services via Traefik
	if err := createHomelabForwardAuth(ac); err != nil {
		return err
	}

	return nil
}

// createHomelabForwardAuth creates the Authentik proxy provider + embedded outpost
// that backs Traefik's ForwardAuth middleware for public K8s services (grafana, harbor, etc.).
func createHomelabForwardAuth(ac AuthentikContext) error {
	proxyProvider, err := authentik.NewProviderProxy(ac.Ctx, "provider-homelab-forwardauth", &authentik.ProviderProxyArgs{
		Name:              pulumi.String("Homelab ForwardAuth"),
		Mode:              pulumi.String("forward_domain"),
		ExternalHost:      pulumi.String("https://madhan.app"),
		CookieDomain:      pulumi.String(".madhan.app"),
		AuthorizationFlow: ac.FlowImplicit,
		InvalidationFlow:  ac.FlowInvalid,
	}, pulumi.Provider(ac.Provider))
	if err != nil {
		return err
	}

	_, err = authentik.NewApplication(ac.Ctx, "app-homelab-forwardauth", &authentik.ApplicationArgs{
		Name: pulumi.String("Homelab ForwardAuth"),
		Slug: pulumi.String("homelab-forwardauth"),
		ProtocolProvider: proxyProvider.ID().ApplyT(func(id pulumi.ID) (*float64, error) {
			idInt, err := strconv.Atoi(string(id))
			if err != nil {
				return nil, err
			}
			f := float64(idInt)
			return &f, nil
		}).(pulumi.Float64PtrOutput),
	}, pulumi.Provider(ac.Provider))
	if err != nil {
		return err
	}

	proxyProviderID := proxyProvider.ID().ApplyT(func(id pulumi.ID) (float64, error) {
		idInt, err := strconv.Atoi(string(id))
		if err != nil {
			return 0, err
		}
		return float64(idInt), nil
	}).(pulumi.Float64Output)

	_, err = authentik.NewOutpost(ac.Ctx, "outpost-homelab-embedded", &authentik.OutpostArgs{
		Name: pulumi.String("Homelab Embedded Outpost"),
		Type: pulumi.String("proxy"),
		ProtocolProviders: pulumi.Float64Array{
			proxyProviderID,
		},
	}, pulumi.Provider(ac.Provider))
	return err
}

func createOIDCApp(ac AuthentikContext, app OIDCApp) error {
	fmt.Printf("DEBUG: Creating OIDC App -> Name: %s, Slug: %s, ClientID: %s, ClientType: %s\n", app.Name, app.Slug, app.ClientID, app.ClientType)
	fmt.Printf("DEBUG: LaunchURL: %s\n", app.LaunchURL)
	fmt.Printf("DEBUG: Redirects: %v\n", app.Redirects)

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

	// Default to public if not set
	clientType := app.ClientType
	if clientType == "" {
		clientType = "public"
	}

	providerArgs := &authentik.ProviderOauth2Args{
		Name:                  pulumi.String(app.Name),
		ClientId:              pulumi.String(app.ClientID),
		ClientType:            pulumi.String(clientType),
		AuthorizationFlow:     ac.FlowExplicit,
		InvalidationFlow:      ac.FlowInvalid,
		PropertyMappings:      ac.Scopes,
		AllowedRedirectUris:   redirectMap,
		SubMode:               pulumi.String("user_id"),
		IncludeClaimsInIdToken: pulumi.Bool(true),
		// IMPORTANT: Signing Key is required for verifying JWTs in OIDC (especially for Netbird)
		SigningKey: ac.SigningKey,
		// Match NetBird documentation recommendation
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

	providerIDPtr := oidcProvider.ID().ApplyT(func(id pulumi.ID) (float64, error) {
		idInt, err := strconv.Atoi(string(id))
		if err != nil {
			return 0, fmt.Errorf("failed to convert ID: %w", err)
		}
		return float64(idInt), nil
	}).(pulumi.Float64Output).ToFloat64PtrOutput()

	_, err = authentik.NewApplication(ac.Ctx, fmt.Sprintf("app-%s", app.ClientID), &authentik.ApplicationArgs{
		Name:             pulumi.String(app.Name),
		Slug:             pulumi.String(app.Slug),
		ProtocolProvider: providerIDPtr,
		MetaLaunchUrl:    pulumi.String(app.LaunchURL),
	}, pulumi.Provider(ac.Provider))

	return err
}

func mustLookupFlow(ctx *pulumi.Context, prov *authentik.Provider, slug string) pulumi.StringInput {
	flow, err := authentik.LookupFlow(ctx, &authentik.LookupFlowArgs{
		Slug: pulumi.StringRef(slug),
	}, pulumi.Provider(prov))
	if err != nil {
		panic(fmt.Sprintf("CRITICAL: Missing flow '%s': %v", slug, err))
	}
	return pulumi.String(flow.Id)
}

func getCommonScopes(ctx *pulumi.Context, prov *authentik.Provider) pulumi.StringArray {
	var scopeIds pulumi.StringArray

	names := []string{"email", "profile", "openid", "offline_access"}

	for _, name := range names {
		scope, err := authentik.LookupPropertyMappingProviderScope(ctx, &authentik.LookupPropertyMappingProviderScopeArgs{
			Managed: pulumi.StringRef(fmt.Sprintf("goauthentik.io/providers/oauth2/scope-%s", name)),
		}, pulumi.Provider(prov))
		if err != nil {
			panic(fmt.Sprintf("CRITICAL: Missing scope '%s'", name))
		}
		scopeIds = append(scopeIds, pulumi.String(scope.Id))
	}

	return scopeIds
}
