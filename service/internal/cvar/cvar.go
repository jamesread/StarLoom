package cvar

const (
	TypeString = "string"
	TypeInt    = "int"
	TypeBool   = "bool"
)

const (
	KeySiteTitle                = "site_title"
	KeyShowFooter               = "show_footer"
	KeyShowVersionNumber        = "show_version_number"
	KeyShowNewVersions          = "show_new_versions"
	KeyDefaultAwardStars        = "default_award_stars"
	KeyEnableRedemptionApproval = "enable_redemption_approval"

	KeyThemeColorSchemeSwitcherEnabled = "theme_color_scheme_switcher_enabled"
	KeyThemeName                       = "theme_name"
	KeyThemeControl                    = "theme_control"

	ThemeControlSystem = "system"
	ThemeControlUser   = "user"

	DefaultAwardStars = 1
	MaxAwardStars     = 100

	CategorySite     = "Site"
	CategoryFeatures = "Features"
	CategoryTheme    = "Theme"
)

type Def struct {
	Key         string
	MainType    string
	Title       string
	Description string
	Category    string
	ValueString string
	Ordinal     int
	ValueInt    int
}

// Defaults returns the cvar catalog. siteTitle and showFooter seed first insert only.
func Defaults(siteTitle string, showFooter bool) []Def {
	showFooterInt := 0
	if showFooter {
		showFooterInt = 1
	}
	if siteTitle == "" {
		siteTitle = "StarApp"
	}

	return []Def{
		{
			Key: KeySiteTitle, MainType: TypeString, ValueString: siteTitle,
			Title: "Site title", Description: "Shown in the header and browser tab.",
			Category: CategorySite, Ordinal: 10,
		},
		{
			Key: KeyShowFooter, MainType: TypeBool, ValueInt: showFooterInt,
			Title: "Show footer", Description: "Display the page footer with version and links.",
			Category: CategorySite, Ordinal: 20,
		},
		{
			Key: KeyShowVersionNumber, MainType: TypeBool, ValueInt: 1,
			Title: "Show version number", Description: "Display the installed version in the footer.",
			Category: CategorySite, Ordinal: 25,
		},
		{
			Key: KeyShowNewVersions, MainType: TypeBool, ValueInt: 1,
			Title: "Show new versions", Description: "Offer a link when a newer release is available.",
			Category: CategorySite, Ordinal: 26,
		},
		{
			Key: KeyDefaultAwardStars, MainType: TypeInt, ValueInt: DefaultAwardStars,
			Title: "Default award stars", Description: "Default number of stars when a parent awards without entering a amount (1–100).",
			Category: CategoryFeatures, Ordinal: 30,
		},
		{
			Key: KeyEnableRedemptionApproval, MainType: TypeBool, ValueInt: 1,
			Title:       "Redemption approval by default",
			Description: "When enabled, new rewards require parent approval before stars are deducted.",
			Category:    CategoryFeatures, Ordinal: 40,
		},
		{
			Key: KeyThemeColorSchemeSwitcherEnabled, MainType: TypeBool, ValueInt: 0,
			Title:       "Color scheme switcher",
			Description: "Show the auto/light/dark color scheme control in the PicoCrank header.",
			Category:    CategoryTheme, Ordinal: 10,
		},
		{
			Key: KeyThemeName, MainType: TypeString, ValueString: "",
			Title:       "Theme name",
			Description: "Default or enforced drop-in CSS theme (empty = Femtocrank base styling only).",
			Category:    CategoryTheme, Ordinal: 20,
		},
		{
			Key: KeyThemeControl, MainType: TypeString, ValueString: ThemeControlUser,
			Title:       "Theme control",
			Description: "System preference forces the theme name cvar for everyone. User preference uses the cvar as default and allows overrides on User Preferences.",
			Category:    CategoryTheme, Ordinal: 30,
		},
	}
}
