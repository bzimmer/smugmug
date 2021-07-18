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
	NiceName            string     `json:"NiceName"`
	URLName             string     `json:"UrlName"`
	Title               string     `json:"Title"`
	Name                string     `json:"Name"`
	AllowDownloads      bool       `json:"AllowDownloads"`
	Description         string     `json:"Description"`
	EXIF                bool       `json:"EXIF"`
	External            bool       `json:"External"`
	Filenames           bool       `json:"Filenames"`
	Geography           bool       `json:"Geography"`
	Keywords            string     `json:"Keywords"`
	PasswordHint        string     `json:"PasswordHint"`
	Protected           bool       `json:"Protected"`
	SortDirection       string     `json:"SortDirection"`
	SortMethod          string     `json:"SortMethod"`
	SecurityType        string     `json:"SecurityType"`
	CommerceLightbox    bool       `json:"CommerceLightbox"`
	AlbumKey            string     `json:"AlbumKey"`
	CanBuy              bool       `json:"CanBuy"`
	CanFavorite         bool       `json:"CanFavorite"`
	LastUpdated         *time.Time `json:"LastUpdated"`
	ImagesLastUpdated   *time.Time `json:"ImagesLastUpdated"`
	NodeID              string     `json:"NodeID"`
	ImageCount          int        `json:"ImageCount"`
	URLPath             string     `json:"UrlPath"`
	CanShare            bool       `json:"CanShare"`
	HasDownloadPassword bool       `json:"HasDownloadPassword"`
	Packages            bool       `json:"Packages"`
	URI                 string     `json:"Uri"`
	WebURI              string     `json:"WebUri"`
	URIDescription      string     `json:"UriDescription"`
	URIs                struct {
		AlbumShareURIs           *APIEndpoint `json:"AlbumShareUris"`
		Node                     *APIEndpoint `json:"Node"`
		NodeCoverImage           *APIEndpoint `json:"NodeCoverImage"`
		User                     *APIEndpoint `json:"User"`
		Folder                   *APIEndpoint `json:"Folder"`
		ParentFolders            *APIEndpoint `json:"ParentFolders"`
		HighlightImage           *APIEndpoint `json:"HighlightImage"`
		AlbumHighlightImage      *APIEndpoint `json:"AlbumHighlightImage"`
		AlbumImages              *APIEndpoint `json:"AlbumImages"`
		AlbumPopularMedia        *APIEndpoint `json:"AlbumPopularMedia"`
		AlbumGeoMedia            *APIEndpoint `json:"AlbumGeoMedia"`
		AlbumDownload            *APIEndpoint `json:"AlbumDownload"`
		AlbumPrices              *APIEndpoint `json:"AlbumPrices"`
		AlbumPricelistExclusions *APIEndpoint `json:"AlbumPricelistExclusions"`
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
	Title            string     `json:"Title"`
	Caption          string     `json:"Caption"`
	Keywords         string     `json:"Keywords"`
	KeywordArray     []string   `json:"KeywordArray"`
	Watermark        string     `json:"Watermark"`
	Latitude         Coordinate `json:"Latitude"`
	Longitude        Coordinate `json:"Longitude"`
	Altitude         int        `json:"Altitude"`
	Hidden           bool       `json:"Hidden"`
	ThumbnailURL     string     `json:"ThumbnailUrl"`
	FileName         string     `json:"FileName"`
	Processing       bool       `json:"Processing"`
	UploadKey        string     `json:"UploadKey"`
	Date             *time.Time `json:"Date"`
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
		Caption struct {
			HTML string `json:"html"`
			Text string `json:"text"`
		} `json:"Caption"`
		FileName struct {
			HTML string `json:"html"`
			Text string `json:"text"`
		} `json:"FileName"`
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
	Request struct {
		Version string `json:"Version"`
		Method  string `json:"Method"`
		URI     string `json:"Uri"`
	} `json:"Request"`
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

	URI            string `json:"Uri,omitempty"`
	URIDescription string `json:"UriDescription,omitempty"`
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

	URI            string `json:"Uri,omitempty"`
	URIDescription string `json:"UriDescription,omitempty"`
}
