package cvar

const (
	TypeString = "string"
	TypeInt    = "int"
	TypeBool   = "bool"
)

const (
	KeySiteTitle                = "site_title"
	KeyShowFooter               = "show_footer"
	KeyDefaultAwardStars        = "default_award_stars"
	KeyEnableRedemptionApproval = "enable_redemption_approval"

	DefaultAwardStars = 1
	MaxAwardStars     = 100

	CategorySite     = "Site"
	CategoryFeatures = "Features"
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
	}
}
