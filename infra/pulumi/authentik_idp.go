package main

import (
	"fmt"
	"strconv"
	"strings"

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

type AuthentikContext struct {
	Ctx          *pulumi.Context
	Provider     *authentik.Provider
	FlowAuth     pulumi.StringInput
	FlowInvalid  pulumi.StringInput
	FlowImplicit pulumi.StringInput
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
	token := k.String("AUTHENTIK_TOKEN")
	ghSecret := k.String("AUTHENTIK_GITHUB_SECRET")
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
	_, err = authentik.NewSourceOauth(ctx, "source-github", &authentik.SourceOauthArgs{
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

	// --- Applications ---

	// Netbird (SPA -> Public)
	// Requires a valid Signing Key for JWT verification
	if err := createOIDCApp(ac, OIDCApp{
		Name:       "Netbird",
		Slug:       "netbird",
		ClientID:   "aumenijDycfG1cQURqH9BNJpV3KVUCoMHGPUVUlT",
		ClientType: "public",
		LaunchURL:  "https://netbird.madhan.app/",
		Redirects: []string{
			"https://netbird.madhan.app/.*",
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
		Name:                pulumi.String(app.Name),
		ClientId:            pulumi.String(app.ClientID),
		ClientType:          pulumi.String(clientType),
		AuthorizationFlow:   ac.FlowAuth,
		InvalidationFlow:    ac.FlowInvalid,
		PropertyMappings:    ac.Scopes,
		AllowedRedirectUris: redirectMap,
		SubMode:             pulumi.String("user_id"),
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
