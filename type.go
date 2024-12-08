package nyckell

import (
	"bytes"
	"fmt"
	"net/url"
	"time"
)

var nyckelClientId = envVarOrPanic("NYKEL_CLIENT_ID")
var nyckelClientSecret = envVarOrPanic("NYKEL_CLIENT_SECRET")

const AccessTokenEndpoint = "https://www.nyckel.com/connect/token"
const ListFunctionsEndpoint = "https://www.nyckel.com/v1/functions"
const FindFunctionsEndpoint = "https://www.nyckel.com/v1/functions/%s"
const CreateFunctionsEndpoint = "https://www.nyckel.com/v1/functions"
const ListLabelsEndpoint = "https://www.nyckel.com/v1/functions/%s/labels"
const DeleteLabelEndpoint = "https://www.nyckel.com/v1/functions/%s/labels/%s"
const CreateLabelEndpoint = "https://www.nyckel.com/v1/functions/%s/labels"
const CreateImageSampleEndpoint = "https://www.nyckel.com/v1/functions/%s/samples"
const AnnotateSampleEndpoint = "https://www.nyckel.com/v1/functions/%s/samples/%s/annotation"
const ListSamplesEndpoint = "https://www.nyckel.com/v1/functions/%s/samples"
const DeleteSampleEndpoint = "https://www.nyckel.com/v1/functions/%s/samples/%s"
const InvokeFunctionEndpoint = "https://www.nyckel.com/v1/functions/%s/invoke"

var _ = NyckelInput(&AccessTokenIn{})

type NyckelAPI interface {
	AccessToken(now time.Time) (*AccessToken, error)
	ListFunction() ([]Function, error)
	FindFunctionById(id string) (Function, error)
	CreateFunction(name string, in FunctionInputType, out FunctionOutputType) (Function, error)
	ListLabels(funcId string) ([]Label, error)
	DeleteLabel(funcId string, labelId string) error
	CreateLabel(funcId string, name string, description string) (Label, error)
	CreateImageSample(funcId string, fileName string, labelName string, ourId string) (*ImageSample, error)
	CreateImageSampleId(funcId string, fileName string, labelId string, ourId string) (*ImageSample, error)
	ListSamples(funcid string, count int, start int, end int, externalId string) ([]Sample, error)
	DeleteSample(funcId string, sampleId string) error
	AnnotateSample(funcId string, sampleId string, labelName string) (string, error)
	AnnotateSampleId(funcId string, sampleId string, labelId string) (string, error)
	InvokeFunction(funcId string, fileName string, capture bool, ourId string) (Invocation, error)
}

//
// AccessToken
//

type AccessTokenIn struct {
}

type AccessToken struct {
	Creation    time.Time
	AccessToken string  `json:"access_token"`
	ExpiresRaw  float64 `json:"expires_in"`
	ExpiresSec  int
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

// ErrNyckel is the parsed body of a response that failed.
type ErrNyckel struct {
	Text string `json:"error"`
}

// NyckelCallResult is the raw material returned from an API call.
type NyckelCallResult struct {
	StatusCode int
	Result     any           // nil on failure
	RawError   error         //nil on success
	Buffer     *bytes.Buffer // on success or failure of the api (it had a body) non-nil
}

func (n *NyckelCallResult) Error() string {
	if n.RawError == nil {
		return ""
	}
	return n.RawError.Error()
}

// NykellInput is an interface that represents an entity that does not need auth.
type NyckelInput interface {
	FormData() url.Values
}

// NyckelAPIDef is the struct that represents the api.
type NyckelAPIDef struct {
	accessToken  *AccessToken
	panicOnError bool
}

//
// Function
//

type Function struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Input  string `json:"input"`
	Output string `json:"output"`
}

//
// Function I/O types
//

type FunctionInputType int
type FunctionOutputType int

const (
	FunctionInputTypeText    FunctionInputType = 1
	FunctionInputTypeImage   FunctionInputType = 2
	FunctionInputTypeTabular FunctionInputType = 3

	FunctionOutputTypeClassification FunctionOutputType = 1
	FunctionOutputTypeTags           FunctionOutputType = 2
	FunctionOutputTypeSearch         FunctionOutputType = 3
	FunctionOutputTypeLocalization   FunctionOutputType = 4
	FunctionOutputTypeOCR            FunctionOutputType = 5
)

func (f FunctionInputType) String() string {
	switch f {
	case 1:
		return "Text"
	case 2:
		return "Image"
	case 3:
		return "Tabular"
	default:
		panic(fmt.Sprintf("unknown or unset function input type: %d", f))
	}
}
func (f FunctionOutputType) String() string {
	switch f {
	case 1:
		return "Classification"
	case 2:
		return "Tags"
	case 3:
		return "Search"
	case 4:
		return "Localization"
	case 5:
		return "OCR"
	default:
		panic(fmt.Sprintf("unknown function output type: %d", f))
	}
}

//
// Label
//

type Label struct {
	Id          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MetaData    string `json:"metadata,omitempty"`
}

// Image Sample
type ImageSample struct {
	Id         string     `json:"id"`
	Data       string     `json:"data"`
	ExternalId string     `json:"externalId"`
	Annotation Annotation `json:"annotation"`
}

//
// Sample
//

type Sample struct {
	Id         string     `json:"id"`
	Data       string     `json:"data"`
	Annotation Annotation `json:"annotation"`
	Prediction Prediction `json:"prediction"`
	ExternalId string     `json:"externalId"`
}

//
// Annotation
//

type Annotation struct {
	LabelId string `json:"labelId"`
}

//
// Prediction
//

type Prediction struct {
	LabelId    string  `json:"labelId"`
	Confidence float64 `json:"confidence"`
}

//
// Invocation
//

type Invocation struct {
	LabelName  string  `json:"labelName"`
	LabelId    string  `json:"labelId"`
	Confidence float64 `json:"confidence"`
}
