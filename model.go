package smugmug

import (
	"encoding/json"
	"io"
	"strconv"
	"time"
)

type Fault struct { //nolint:errname // smugmug naming convention
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (f *Fault) Error() string {
	return f.Message
}

type Coordinate float64

// UnmarshalJSON converts the json value to a coordinate
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

type FormattedValue struct {
	HTML string `json:"html"`
	Text string `json:"text"`
}

type FormattedValues struct {
	Name        *FormattedValue `json:"Name"`
	Description *FormattedValue `json:"Description"`
}

type APIEndpoint struct {
	URI            string `json:"Uri"`
	Locator        string `json:"Locator"`
	LocatorType    string `json:"LocatorType"`
	URIDescription string `json:"UriDescription"`
	EndpointType   string `json:"EndpointType"`
}

type UserURIs struct {
	// Folder             *APIEndpoint `json:"Folder"`
	BioImage           *APIEndpoint `json:"BioImage"`
	CoverImage         *APIEndpoint `json:"CoverImage"`
	Features           *APIEndpoint `json:"Features"`
	Node               *APIEndpoint `json:"Node"`
	SiteSettings       *APIEndpoint `json:"SiteSettings"`
	URLPathLookup      *APIEndpoint `json:"UrlPathLookup"`
	UserAlbums         *APIEndpoint `json:"UserAlbums"`
	UserFeaturedAlbums *APIEndpoint `json:"UserFeaturedAlbums"`
	UserGeoMedia       *APIEndpoint `json:"UserGeoMedia"`
	UserImageSearch    *APIEndpoint `json:"UserImageSearch"`
	UserPopularMedia   *APIEndpoint `json:"UserPopularMedia"`
	UserProfile        *APIEndpoint `json:"UserProfile"`
	UserRecentImages   *APIEndpoint `json:"UserRecentImages"`
	UserTopKeywords    *APIEndpoint `json:"UserTopKeywords"`
}

type User struct {
	NickName       string   `json:"NickName"`
	ViewPassHint   string   `json:"ViewPassHint"`
	RefTag         string   `json:"RefTag"`
	Name           string   `json:"Name"`
	QuickShare     bool     `json:"QuickShare"`
	URI            string   `json:"Uri"`
	WebURI         string   `json:"WebUri"`
	URIDescription string   `json:"UriDescription"`
	URIs           UserURIs `json:"Uris"`
	ResponseLevel  string   `json:"ResponseLevel"`
	// expansions
	Node *Node `json:"Node"`
}

type AlbumURIs struct {
	// Folder                     *APIEndpoint `json:"Folder"`
	// ParentFolders              *APIEndpoint `json:"ParentFolders"`
	AlbumShareUris             *APIEndpoint `json:"AlbumShareUris"`
	Node                       *APIEndpoint `json:"Node"`
	NodeCoverImage             *APIEndpoint `json:"NodeCoverImage"`
	User                       *APIEndpoint `json:"User"`
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
	URIs                   AlbumURIs  `json:"Uris"`
	ResponseLevel          string     `json:"ResponseLevel"`
	// expansions
	User           *User  `json:"User"`
	Node           *Node  `json:"Node"`
	HighlightImage *Image `json:"HighlightImage"`
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

type ImageURIs struct {
	// Album and ImageAlbum are used in different context but should be identical
	Album                         *APIEndpoint `json:"Album"`
	AlbumImageMetadata            *APIEndpoint `json:"AlbumImageMetadata"`
	AlbumImagePricelistExclusions *APIEndpoint `json:"AlbumImagePricelistExclusions"`
	AlbumImageShareURIs           *APIEndpoint `json:"AlbumImageShareUris"`
	Image                         *APIEndpoint `json:"Image"`
	ImageAlbum                    *APIEndpoint `json:"ImageAlbum"`
	ImageComments                 *APIEndpoint `json:"ImageComments"`
	ImageMetadata                 *APIEndpoint `json:"ImageMetadata"`
	ImagePricelistExclusions      *APIEndpoint `json:"ImagePricelistExclusions"`
	ImagePrices                   *APIEndpoint `json:"ImagePrices"`
	ImageSizeDetails              *APIEndpoint `json:"ImageSizeDetails"`
	ImageSizes                    *APIEndpoint `json:"ImageSizes"`
	LargestImage                  *APIEndpoint `json:"LargestImage"`
	PointOfInterest               *APIEndpoint `json:"PointOfInterest"`
	PointOfInterestCrops          *APIEndpoint `json:"PointOfInterestCrops"`
	Regions                       *APIEndpoint `json:"Regions"`
}

type Image struct {
	// Date             *time.Time `json:"Date"` // deprecated, use DateTimeUploaded
	// Watermark        string     `json:"Watermark"` // deprecated
	Title            string           `json:"Title"`
	Caption          string           `json:"Caption"`
	Keywords         string           `json:"Keywords"`
	KeywordArray     []string         `json:"KeywordArray"`
	Latitude         Coordinate       `json:"Latitude"`
	Longitude        Coordinate       `json:"Longitude"`
	Altitude         int              `json:"Altitude"`
	Hidden           bool             `json:"Hidden"`
	ThumbnailURL     string           `json:"ThumbnailUrl"`
	FileName         string           `json:"FileName"`
	Processing       bool             `json:"Processing"`
	UploadKey        string           `json:"UploadKey"`
	DateTimeUploaded *time.Time       `json:"DateTimeUploaded"`
	DateTimeOriginal *time.Time       `json:"DateTimeOriginal"`
	Format           string           `json:"Format"`
	OriginalHeight   int              `json:"OriginalHeight"`
	OriginalWidth    int              `json:"OriginalWidth"`
	OriginalSize     int              `json:"OriginalSize"`
	LastUpdated      *time.Time       `json:"LastUpdated"`
	Collectable      bool             `json:"Collectable"`
	IsArchive        bool             `json:"IsArchive"`
	IsVideo          bool             `json:"IsVideo"`
	CanEdit          bool             `json:"CanEdit"`
	CanBuy           bool             `json:"CanBuy"`
	Protected        bool             `json:"Protected"`
	ImageKey         string           `json:"ImageKey"`
	Serial           int              `json:"Serial"`
	ArchivedURI      string           `json:"ArchivedUri"`
	ArchivedSize     int              `json:"ArchivedSize"`
	ArchivedMD5      string           `json:"ArchivedMD5"`
	CanShare         bool             `json:"CanShare"`
	Comments         bool             `json:"Comments"`
	ShowKeywords     bool             `json:"ShowKeywords"`
	FormattedValues  *FormattedValues `json:"FormattedValues"`
	URI              string           `json:"Uri"`
	URIDescription   string           `json:"UriDescription"`
	URIs             ImageURIs        `json:"Uris"`
	Movable          bool             `json:"Movable"`
	Origin           string           `json:"Origin"`
	WebURI           string           `json:"WebUri"`
	// expansions
	Album            *Album            `json:"Album"`
	ImageSizeDetails *ImageSizeDetails `json:"ImageSizeDetails"`
}

type ImageSize struct {
	URL    string `json:"Url,omitempty"`
	Ext    string `json:",omitempty"`
	Height int    `json:",omitempty"`
	Width  int    `json:",omitempty"`
	Size   int64  `json:",omitempty"`
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

type NodeURIs struct {
	// FolderByID     *APIEndpoint `json:"FolderByID"`
	Album          *APIEndpoint `json:"Album"`
	Children       *APIEndpoint `json:"ChildNodes"`
	HighlightImage *APIEndpoint `json:"HighlightImage"`
	MoveNodes      *APIEndpoint `json:"MoveNodes"`
	NodeComments   *APIEndpoint `json:"NodeComments"`
	NodeCoverImage *APIEndpoint `json:"NodeCoverImage"`
	NodeGrants     *APIEndpoint `json:"NodeGrants"`
	NodePageDesign *APIEndpoint `json:"NodePageDesign"`
	Parent         *APIEndpoint `json:"ParentNode"`
	Parents        *APIEndpoint `json:"ParentNodes"`
	User           *APIEndpoint `json:"User"`
}

type Nodelet struct {
	Type    string `json:"Type"`
	Name    string `json:"Name"`
	URLName string `json:"UrlName"`
	Privacy string `json:"Privacy"`
}

type Node struct {
	Nodelet
	CoverImageURI         string           `json:"CoverImageUri"`
	Description           string           `json:"Description"`
	HideOwner             bool             `json:"HideOwner"`
	HighlightImageURI     string           `json:"HighlightImageUri"`
	Keywords              []string         `json:"Keywords"`
	Password              string           `json:"Password"`
	PasswordHint          string           `json:"PasswordHint"`
	SecurityType          string           `json:"SecurityType"`
	ShowCoverImage        bool             `json:"ShowCoverImage"`
	SmugSearchable        string           `json:"SmugSearchable"`
	SortDirection         string           `json:"SortDirection"`
	SortMethod            string           `json:"SortMethod"`
	WorldSearchable       string           `json:"WorldSearchable"`
	DateAdded             *time.Time       `json:"DateAdded"`
	DateModified          *time.Time       `json:"DateModified"`
	EffectivePrivacy      string           `json:"EffectivePrivacy"`
	EffectiveSecurityType string           `json:"EffectiveSecurityType"`
	FormattedValues       *FormattedValues `json:"FormattedValues"`
	HasChildren           bool             `json:"HasChildren"`
	IsRoot                bool             `json:"IsRoot"`
	NodeID                string           `json:"NodeID"`
	URLPath               string           `json:"UrlPath"`
	URI                   string           `json:"Uri"`
	WebURI                string           `json:"WebUri"`
	URIDescription        string           `json:"UriDescription"`
	URIs                  NodeURIs         `json:"Uris"`
	ResponseLevel         string           `json:"ResponseLevel"`
	// expansions
	User           *User  `json:"User"`
	Album          *Album `json:"Album"`
	Parent         *Node  `json:"Parent"`
	HighlightImage *Image `json:"HighlightImage"`
}

// Uploadable holds the details about an image suitable for upload
type Uploadable struct {
	// Name is the basename of the image (not the full path)
	Name string `json:"Name"`
	// Size is the size in bytes
	Size int64 `json:"Size"`
	// MD5 is the hash of the file contents
	MD5 string `json:"MD5"`
	// Replaces is the URI of an image to replace
	Replaces string `json:"Replaces"`
	// AlbumKey is the album into which the file will be uploaded
	AlbumKey string `json:"AlbumKey"`
	// Reader holds the image data for uploading
	Reader io.Reader `json:"-"`
}

// Upload is the object details for the uploaded object
type Upload struct {
	// Status of the request
	Status string `json:"Status"`
	// Method is action performed
	Method string `json:"Method"`
	// ImageURI is the uri of the object
	ImageURI string `json:"ImageUri"`
	// Elapsed time of the upload
	Elapsed time.Duration `json:"Elapsed"`
	// AlbumImageURI is the uri of the object in the album
	AlbumImageURI string `json:"AlbumImageUri"`
	// URL is the url of the uploaded object
	URL string `json:"URL"`
	// Uploadable is the object being uploaded
	Uploadable *Uploadable `json:"Uploadable"`
}
