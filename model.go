package smugmug

import (
	"encoding/json"
	"strconv"
	"time"
)

type Fault struct {
	Message string
}

func (f *Fault) Error() string {
	return f.Message
}

type Coordinate float64

func (c *Coordinate) UnmarshalJSON(b []byte) error {
	var s interface{}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch x := s.(type) {
	case float64:
		*c = Coordinate(x)
	case string:
		f, err := strconv.ParseFloat(x, 64)
		if err != nil {
			return err
		}
		*c = Coordinate(f)
	}
	return nil
}

type Request struct {
	Version string `json:"Version"`
	Method  string `json:"Method"`
	URI     string `json:"Uri"`
}

type Timing struct {
	Total struct {
		Time    float64 `json:"time"`
		Cycles  int     `json:"cycles"`
		Objects int     `json:"objects"`
	} `json:"Total"`
}

type FormattedValue struct {
	HTML string `json:"html"`
	Text string `json:"text"`
}

type Options struct {
	MethodDetails struct {
		OPTIONS struct {
			Permissions []string `json:"Permissions"`
		} `json:"OPTIONS"`
		GET struct {
			Permissions []string `json:"Permissions"`
		} `json:"GET"`
	} `json:"MethodDetails"`
	Methods    []string `json:"Methods"`
	MediaTypes []string `json:"MediaTypes"`
	Output     []struct {
		Name        string   `json:"Name"`
		Type        string   `json:"Type"`
		OPTIONS     []string `json:"OPTIONS,omitempty"`
		MINCOUNT    int      `json:"MIN_COUNT,omitempty"`
		MAXCOUNT    int      `json:"MAX_COUNT,omitempty"`
		MINCHARS    int      `json:"MIN_CHARS,omitempty"`
		MAXCHARS    int      `json:"MAX_CHARS,omitempty"`
		MINVALUE    int      `json:"MIN_VALUE,omitempty"`
		MAXVALUE    int      `json:"MAX_VALUE,omitempty"`
		Description string   `json:"Description,omitempty"`
	} `json:"Output"`
	ResponseLevels []string `json:"ResponseLevels"`
	Path           []struct {
		Type       string `json:"type"`
		Text       string `json:"text,omitempty"`
		ParamName  string `json:"param_name,omitempty"`
		ParamValue string `json:"param_value,omitempty"`
	} `json:"Path"`
}

type APIEndpoint struct {
	URI            string `json:"Uri"`
	Locator        string `json:"Locator"`
	LocatorType    string `json:"LocatorType"`
	URIDescription string `json:"UriDescription"`
	EndpointType   string `json:"EndpointType"`
}

type User struct {
	NickName       string `json:"NickName"`
	ViewPassHint   string `json:"ViewPassHint"`
	RefTag         string `json:"RefTag"`
	Name           string `json:"Name"`
	QuickShare     bool   `json:"QuickShare"`
	URI            string `json:"Uri"`
	WebURI         string `json:"WebUri"`
	URIDescription string `json:"UriDescription"`
	URIs           struct {
		BioImage           *APIEndpoint `json:"BioImage"`
		CoverImage         *APIEndpoint `json:"CoverImage"`
		UserProfile        *APIEndpoint `json:"UserProfile"`
		Node               *APIEndpoint `json:"Node"`
		Folder             *APIEndpoint `json:"Folder"`
		Features           *APIEndpoint `json:"Features"`
		SiteSettings       *APIEndpoint `json:"SiteSettings"`
		UserAlbums         *APIEndpoint `json:"UserAlbums"`
		UserGeoMedia       *APIEndpoint `json:"UserGeoMedia"`
		UserPopularMedia   *APIEndpoint `json:"UserPopularMedia"`
		UserFeaturedAlbums *APIEndpoint `json:"UserFeaturedAlbums"`
		UserRecentImages   *APIEndpoint `json:"UserRecentImages"`
		UserImageSearch    *APIEndpoint `json:"UserImageSearch"`
		UserTopKeywords    *APIEndpoint `json:"UserTopKeywords"`
		URLPathLookup      *APIEndpoint `json:"UrlPathLookup"`
	} `json:"Uris"`
	ResponseLevel string          `json:"ResponseLevel"`
	Expansions    *UserExpansions `json:"-"`
}

type UserExpansions struct {
	Albums []*Album `json:"-"`
}

type UserResponse struct {
	Request  *Request `json:"Request"`
	Options  *Options `json:"Options"`
	Response struct {
		URI            string  `json:"Uri"`
		Locator        string  `json:"Locator"`
		LocatorType    string  `json:"LocatorType"`
		User           *User   `json:"User"`
		URIDescription string  `json:"UriDescription"`
		EndpointType   string  `json:"EndpointType"`
		DocURI         string  `json:"DocUri"`
		Timing         *Timing `json:"Timing"`
	} `json:"Response"`
	Expansions map[string]*json.RawMessage `json:"Expansions,omitempty"`
	Code       int                         `json:"Code"`
	Message    string                      `json:"Message"`
}

type Album struct {
	// NiceName            string     `json:"NiceName"` // deprecated, use URLName
	// Title               string     `json:"Title"` // deprecated, use Name
	URLName                string     `json:"UrlName"`
	Name                   string     `json:"Name"`
	TemplateURI            string     `json:"TemplateUri"`
	AllowDownloads         bool       `json:"AllowDownloads"`
	Backprinting           string     `json:"Backprinting"`
	BoutiquePackaging      string     `json:"BoutiquePackaging"`
	CanRank                bool       `json:"CanRank"`
	Clean                  bool       `json:"Clean"`
	Comments               bool       `json:"Comments"`
	Description            string     `json:"Description"`
	EXIF                   bool       `json:"EXIF"`
	External               bool       `json:"External"`
	FamilyEdit             bool       `json:"FamilyEdit"`
	Filenames              bool       `json:"Filenames"`
	FriendEdit             bool       `json:"FriendEdit"`
	Geography              bool       `json:"Geography"`
	Header                 string     `json:"Header"`
	HideOwner              bool       `json:"HideOwner"`
	InterceptShipping      string     `json:"InterceptShipping"`
	Keywords               string     `json:"Keywords"`
	LargestSize            string     `json:"LargestSize"`
	PackagingBranding      bool       `json:"PackagingBranding"`
	Password               string     `json:"Password"`
	PasswordHint           string     `json:"PasswordHint"`
	Printable              bool       `json:"Printable"`
	Privacy                string     `json:"Privacy"`
	ProofDays              int        `json:"ProofDays"`
	ProofDigital           bool       `json:"ProofDigital"`
	Protected              bool       `json:"Protected"`
	Share                  bool       `json:"Share"`
	Slideshow              bool       `json:"Slideshow"`
	SmugSearchable         string     `json:"SmugSearchable"`
	SortDirection          string     `json:"SortDirection"`
	SortMethod             string     `json:"SortMethod"`
	SquareThumbs           bool       `json:"SquareThumbs"`
	Watermark              bool       `json:"Watermark"`
	WorldSearchable        bool       `json:"WorldSearchable"`
	SecurityType           string     `json:"SecurityType"`
	CommerceLightbox       bool       `json:"CommerceLightbox"`
	HighlightAlbumImageURI string     `json:"HighlightAlbumImageUri"`
	AlbumKey               string     `json:"AlbumKey"`
	CanBuy                 bool       `json:"CanBuy"`
	CanFavorite            bool       `json:"CanFavorite"`
	Date                   *time.Time `json:"Date"`
	LastUpdated            *time.Time `json:"LastUpdated"`
	ImagesLastUpdated      *time.Time `json:"ImagesLastUpdated"`
	NodeID                 string     `json:"NodeID"`
	OriginalSizes          int        `json:"OriginalSizes"`
	TotalSizes             int        `json:"TotalSizes"`
	ImageCount             int        `json:"ImageCount"`
	URLPath                string     `json:"UrlPath"`
	CanShare               bool       `json:"CanShare"`
	HasDownloadPassword    bool       `json:"HasDownloadPassword"`
	Packages               bool       `json:"Packages"`
	URI                    string     `json:"Uri"`
	WebURI                 string     `json:"WebUri"`
	URIDescription         string     `json:"UriDescription"`
	URIs                   struct {
		AlbumShareUris             *APIEndpoint `json:"AlbumShareUris"`
		Node                       *APIEndpoint `json:"Node"`
		NodeCoverImage             *APIEndpoint `json:"NodeCoverImage"`
		User                       *APIEndpoint `json:"User"`
		Folder                     *APIEndpoint `json:"Folder"`
		ParentFolders              *APIEndpoint `json:"ParentFolders"`
		HighlightImage             *APIEndpoint `json:"HighlightImage"`
		AddSamplePhotos            *APIEndpoint `json:"AddSamplePhotos"`
		AlbumHighlightImage        *APIEndpoint `json:"AlbumHighlightImage"`
		AlbumImages                *APIEndpoint `json:"AlbumImages"`
		AlbumPopularMedia          *APIEndpoint `json:"AlbumPopularMedia"`
		AlbumGeoMedia              *APIEndpoint `json:"AlbumGeoMedia"`
		AlbumComments              *APIEndpoint `json:"AlbumComments"`
		MoveAlbumImages            *APIEndpoint `json:"MoveAlbumImages"`
		CollectImages              *APIEndpoint `json:"CollectImages"`
		ApplyAlbumTemplate         *APIEndpoint `json:"ApplyAlbumTemplate"`
		DeleteAlbumImages          *APIEndpoint `json:"DeleteAlbumImages"`
		UploadFromExternalResource *APIEndpoint `json:"UploadFromExternalResource"`
		UploadFromURI              *APIEndpoint `json:"UploadFromUri"`
		AlbumGrants                *APIEndpoint `json:"AlbumGrants"`
		AlbumDownload              *APIEndpoint `json:"AlbumDownload"`
		AlbumPrices                *APIEndpoint `json:"AlbumPrices"`
		AlbumPricelistExclusions   *APIEndpoint `json:"AlbumPricelistExclusions"`
	} `json:"Uris"`
	ResponseLevel string           `json:"ResponseLevel"`
	Expansions    *AlbumExpansions `json:"-"`
}

type AlbumExpansions struct {
	HighlightImage *Image   `json:"HighlightImage"`
	Images         []*Image `json:"Images"`
}

type Pages struct {
	Total          int    `json:"Total"`
	Start          int    `json:"Start"`
	Count          int    `json:"Count"`
	RequestedCount int    `json:"RequestedCount"`
	FirstPage      string `json:"FirstPage"`
	LastPage       string `json:"LastPage"`
	NextPage       string `json:"NextPage"`
}

type AlbumsResponse struct {
	Request  *Request `json:"Request"`
	Options  *Options `json:"Options"`
	Response struct {
		URI            string   `json:"Uri"`
		Locator        string   `json:"Locator"`
		LocatorType    string   `json:"LocatorType"`
		Album          []*Album `json:"Album"`
		URIDescription string   `json:"UriDescription"`
		EndpointType   string   `json:"EndpointType"`
		Pages          *Pages   `json:"Pages"`
		Timing         *Timing  `json:"Timing"`
	} `json:"Response"`
	Expansions map[string]*json.RawMessage `json:",omitempty"`
	Code       int                         `json:"Code"`
	Message    string                      `json:"Message"`
}

type AlbumResponse struct {
	Response struct {
		URI            string  `json:"Uri"`
		Locator        string  `json:"Locator"`
		LocatorType    string  `json:"LocatorType"`
		Album          *Album  `json:"Album"`
		URIDescription string  `json:"UriDescription"`
		EndpointType   string  `json:"EndpointType"`
		DocURI         string  `json:"DocUri"`
		Timing         *Timing `json:"Timing"`
	} `json:"Response"`
	Expansions map[string]*json.RawMessage `json:",omitempty"`
	Code       int                         `json:"Code"`
	Message    string                      `json:"Message"`
}

type Image struct {
	// Date             *time.Time `json:"Date"` // deprecated, use DateTimeUploaded
	// Watermark        string     `json:"Watermark"` // deprecated
	Title            string     `json:"Title"`
	Caption          string     `json:"Caption"`
	Keywords         string     `json:"Keywords"`
	KeywordArray     []string   `json:"KeywordArray"`
	Latitude         Coordinate `json:"Latitude"`
	Longitude        Coordinate `json:"Longitude"`
	Altitude         int        `json:"Altitude"`
	Hidden           bool       `json:"Hidden"`
	ThumbnailURL     string     `json:"ThumbnailUrl"`
	FileName         string     `json:"FileName"`
	Processing       bool       `json:"Processing"`
	UploadKey        string     `json:"UploadKey"`
	DateTimeUploaded *time.Time `json:"DateTimeUploaded"`
	DateTimeOriginal *time.Time `json:"DateTimeOriginal"`
	Format           string     `json:"Format"`
	OriginalHeight   int        `json:"OriginalHeight"`
	OriginalWidth    int        `json:"OriginalWidth"`
	OriginalSize     int        `json:"OriginalSize"`
	LastUpdated      *time.Time `json:"LastUpdated"`
	Collectable      bool       `json:"Collectable"`
	IsArchive        bool       `json:"IsArchive"`
	IsVideo          bool       `json:"IsVideo"`
	CanEdit          bool       `json:"CanEdit"`
	CanBuy           bool       `json:"CanBuy"`
	Protected        bool       `json:"Protected"`
	ImageKey         string     `json:"ImageKey"`
	Serial           int        `json:"Serial"`
	ArchivedURI      string     `json:"ArchivedUri"`
	ArchivedSize     int        `json:"ArchivedSize"`
	ArchivedMD5      string     `json:"ArchivedMD5"`
	CanShare         bool       `json:"CanShare"`
	Comments         bool       `json:"Comments"`
	ShowKeywords     bool       `json:"ShowKeywords"`
	FormattedValues  struct {
		Caption  *FormattedValue `json:"Caption"`
		FileName *FormattedValue `json:"FileName"`
	} `json:"FormattedValues"`
	URI            string `json:"Uri"`
	URIDescription string `json:"UriDescription"`
	URIs           struct {
		LargestImage                  *APIEndpoint `json:"LargestImage"`
		ImageSizes                    *APIEndpoint `json:"ImageSizes"`
		ImageSizeDetails              *APIEndpoint `json:"ImageSizeDetails"`
		PointOfInterest               *APIEndpoint `json:"PointOfInterest"`
		PointOfInterestCrops          *APIEndpoint `json:"PointOfInterestCrops"`
		Regions                       *APIEndpoint `json:"Regions"`
		ImageComments                 *APIEndpoint `json:"ImageComments"`
		ImageMetadata                 *APIEndpoint `json:"ImageMetadata"`
		ImagePrices                   *APIEndpoint `json:"ImagePrices"`
		ImagePricelistExclusions      *APIEndpoint `json:"ImagePricelistExclusions"`
		Album                         *APIEndpoint `json:"Album"`
		Image                         *APIEndpoint `json:"Image"`
		AlbumImagePricelistExclusions *APIEndpoint `json:"AlbumImagePricelistExclusions"`
		AlbumImageMetadata            *APIEndpoint `json:"AlbumImageMetadata"`
		AlbumImageShareUris           *APIEndpoint `json:"AlbumImageShareUris"`
	} `json:"Uris"`
	Movable    bool             `json:"Movable"`
	Origin     string           `json:"Origin"`
	WebURI     string           `json:"WebUri"`
	Expansions *ImageExpansions `json:"-"`
}

type ImageExpansions struct {
	ImageSizeDetails *ImageSizeDetails `json:"ImageSizeDetails"`
}

type ImagesResponse struct {
	Request  *Request `json:"Request"`
	Options  *Options `json:"Options"`
	Response struct {
		URI            string   `json:"Uri"`
		Locator        string   `json:"Locator"`
		LocatorType    string   `json:"LocatorType"`
		Images         []*Image `json:"AlbumImage"`
		URIDescription string   `json:"UriDescription"`
		EndpointType   string   `json:"EndpointType"`
		Pages          *Pages   `json:"Pages"`
		Timing         *Timing  `json:"Timing"`
	} `json:"Response"`
	Expansions map[string]*json.RawMessage `json:",omitempty"`
	Code       int                         `json:"Code"`
	Message    string                      `json:"Message"`
}

type ImageResponse struct {
	Response struct {
		URI            string  `json:"Uri"`
		Locator        string  `json:"Locator"`
		LocatorType    string  `json:"LocatorType"`
		Image          *Image  `json:"Image"`
		URIDescription string  `json:"UriDescription"`
		EndpointType   string  `json:"EndpointType"`
		DocURI         string  `json:"DocUri"`
		Timing         *Timing `json:"Timing"`
	} `json:"Response"`
	Code       int                         `json:"Code"`
	Message    string                      `json:"Message"`
	Expansions map[string]*json.RawMessage `json:",omitempty"`
}

type ImageSize struct {
	URL    string `json:"Url,omitempty"`
	Ext    string `json:",omitempty"`
	Height int    `json:",omitempty"`
	Width  int    `json:",omitempty"`
	Size   int    `json:",omitempty"`
}

type ImageSizeDetails struct {
	ImageSizeLarge    *ImageSize `json:",omitempty"`
	ImageSizeMedium   *ImageSize `json:",omitempty"`
	ImageSizeOriginal *ImageSize `json:",omitempty"`
	ImageSizeSmall    *ImageSize `json:",omitempty"`
	ImageSizeThumb    *ImageSize `json:",omitempty"`
	ImageSizeTiny     *ImageSize `json:",omitempty"`
	ImageSizeX2Large  *ImageSize `json:",omitempty"`
	ImageSizeX3Large  *ImageSize `json:",omitempty"`
	ImageSizeXLarge   *ImageSize `json:",omitempty"`
	ImageURLTemplate  string     `json:"ImageUrlTemplate,omitempty"`
	UsableSizes       []string   `json:",omitempty"`
	URI               string     `json:"Uri,omitempty"`
	URIDescription    string     `json:"UriDescription,omitempty"`
}

type ImageSizes struct {
	LargeImageURL    string `json:"LargeImageUrl,omitempty"`
	LargestImageURL  string `json:"LargestImageUrl,omitempty"`
	MediumImageURL   string `json:"MediumImageUrl,omitempty"`
	OriginalImageURL string `json:"OriginalImageUrl,omitempty"`
	SmallImageURL    string `json:"SmallImageUrl,omitempty"`
	ThumbImageURL    string `json:"ThumbImageUrl,omitempty"`
	TinyImageURL     string `json:"TinyImageUrl,omitempty"`
	X2LargeImageURL  string `json:"X2LargeImageUrl,omitempty"`
	X3LargeImageURL  string `json:"X3LargeImageUrl,omitempty"`
	XLargeImageURL   string `json:"XLargeImageUrl,omitempty"`
	URI              string `json:"Uri,omitempty"`
	URIDescription   string `json:"UriDescription,omitempty"`
}

type Node struct {
	CoverImageURI         string     `json:"CoverImageUri"`
	Description           string     `json:"Description"`
	HideOwner             bool       `json:"HideOwner"`
	HighlightImageURI     string     `json:"HighlightImageUri"`
	Name                  string     `json:"Name"`
	Keywords              []string   `json:"Keywords"`
	Password              string     `json:"Password"`
	PasswordHint          string     `json:"PasswordHint"`
	Privacy               string     `json:"Privacy"`
	SecurityType          string     `json:"SecurityType"`
	ShowCoverImage        bool       `json:"ShowCoverImage"`
	SmugSearchable        string     `json:"SmugSearchable"`
	SortDirection         string     `json:"SortDirection"`
	SortMethod            string     `json:"SortMethod"`
	Type                  string     `json:"Type"`
	URLName               string     `json:"UrlName"`
	WorldSearchable       string     `json:"WorldSearchable"`
	DateAdded             *time.Time `json:"DateAdded"`
	DateModified          *time.Time `json:"DateModified"`
	EffectivePrivacy      string     `json:"EffectivePrivacy"`
	EffectiveSecurityType string     `json:"EffectiveSecurityType"`
	FormattedValues       struct {
		Name        *FormattedValue `json:"Name"`
		Description *FormattedValue `json:"Description"`
	} `json:"FormattedValues"`
	HasChildren    bool   `json:"HasChildren"`
	IsRoot         bool   `json:"IsRoot"`
	NodeID         string `json:"NodeID"`
	URLPath        string `json:"UrlPath"`
	URI            string `json:"Uri"`
	WebURI         string `json:"WebUri"`
	URIDescription string `json:"UriDescription"`
	URIs           struct {
		FolderByID     *APIEndpoint `json:"FolderByID"`
		ParentNode     *APIEndpoint `json:"ParentNode"`
		ParentNodes    *APIEndpoint `json:"ParentNodes"`
		User           *APIEndpoint `json:"User"`
		NodeCoverImage *APIEndpoint `json:"NodeCoverImage"`
		HighlightImage *APIEndpoint `json:"HighlightImage"`
		NodeComments   *APIEndpoint `json:"NodeComments"`
		ChildNodes     *APIEndpoint `json:"ChildNodes"`
		MoveNodes      *APIEndpoint `json:"MoveNodes"`
		NodeGrants     *APIEndpoint `json:"NodeGrants"`
		NodePageDesign *APIEndpoint `json:"NodePageDesign"`
	} `json:"Uris"`
	ResponseLevel string `json:"ResponseLevel"`
}

type NodesResponse struct {
	Request  *Request `json:"Request"`
	Options  *Options `json:"Options"`
	Response struct {
		URI            string  `json:"Uri"`
		Locator        string  `json:"Locator"`
		LocatorType    string  `json:"LocatorType"`
		Node           []*Node `json:"Node"`
		URIDescription string  `json:"UriDescription"`
		EndpointType   string  `json:"EndpointType"`
		Pages          *Pages  `json:"Pages"`
		Timing         *Timing `json:"Timing"`
	} `json:"Response"`
	Expansions map[string]*json.RawMessage `json:",omitempty"`
	Code       int                         `json:"Code"`
	Message    string                      `json:"Message"`
}
