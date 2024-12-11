package nyckell

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var clipLen = 240

// API if you have an AccessToken, if you don't see AccessTokenOrPanic
func NewNyckellAPI(accessToken *AccessToken, panicOnError bool) *NyckelAPIDef {
	return &NyckelAPIDef{accessToken: accessToken, panicOnError: panicOnError}
}

//
// NykelError
//

func (e *ErrNyckel) Error() string {
	return e.Text
}

func NewErrNyckel(s string) *ErrNyckel {
	return &ErrNyckel{Text: s}
}

//
// Panic On Error Flag
//

func (a *NyckelAPIDef) PanicOnError() {
	a.panicOnError = true
}
func (a *NyckelAPIDef) NoPanicOnError() {
	a.panicOnError = false
}

//
// Accesss Token
//

func (a *AccessToken) Expired(now time.Time) bool {
	return a.Creation.Add(time.Duration(a.ExpiresSec) * time.Second).Before(now)
}

func (a *AccessTokenIn) FormData() url.Values {
	formData := url.Values{}
	formData.Add("grant_type", "client_credentials")
	formData.Add("client_id", nyckelClientId)
	formData.Add("client_secret", nyckelClientSecret)
	return formData
}

func accessTokenInput() NyckelInput {
	return &AccessTokenIn{}
}

func (a *AccessTokenIn) Endpoint() string {
	return AccessTokenEndpoint
}

func (a *NyckelAPIDef) AccessTokenOrPanic(now time.Time) *AccessToken {
	result, err := a.AccessToken(now)
	if err != nil {
		panic(err)
	}
	return result
}

func (a *NyckelAPIDef) AccessToken(now time.Time) (*AccessToken, error) {

	token := &AccessToken{}
	callResult := a.getNykelResponseForm((&AccessTokenIn{}).FormData(), token, nil, AccessTokenEndpoint)
	if callResult.RawError != nil {
		return nil, callResult.RawError
	}
	tok := callResult.Result.(*AccessToken)
	tok.Creation = now
	tok.ExpiresSec = int(tok.ExpiresRaw)
	if tok != token {
		panic("insanity")
	}
	return tok, nil
}

//
// Labels
//

func (a *NyckelAPIDef) CreateLabel(funcId string, name string, description string) (Label, error) {
	path := fmt.Sprintf(CreateLabelEndpoint, funcId)
	formData := Label{
		Name:        name,
		Description: description,
	}
	var label Label

	resp := a.getNykelResponseJson(false, formData, &label, a.accessToken, path)
	if resp.RawError != nil {
		if a.panicOnError {
			panic("CreateLabel:" + resp.RawError.Error())
		}
		return Label{}, resp.RawError
	}
	return label, nil
}

func (a *NyckelAPIDef) DeleteLabel(funcId string, labelId string) error {
	path := fmt.Sprintf(DeleteLabelEndpoint, funcId, labelId)
	resp := a.getNyckelResponseRaw(nil, "", nil, "DELETE", a.accessToken, path)
	if resp.RawError != nil {
		if a.panicOnError {
			panic("DeleteLabel: " + resp.RawError.Error())
		}
		return resp.RawError
	}
	return nil
}

func (a *NyckelAPIDef) ListLabels(funcId string) ([]Label, error) {
	collection := []Label{}
	path := fmt.Sprintf(ListLabelsEndpoint, funcId)
	resp := a.getNykelResponseRead(&collection, nil, path)
	if resp.RawError != nil {
		if a.panicOnError {
			panic("ListLabels: " + resp.RawError.Error())
		}
		return nil, resp.RawError
	}
	list := resp.Result.(*[]Label)
	return *list, nil

}

//
// Function
//

func (f *Function) Endpoint() string {
	return ListFunctionsEndpoint
}

func (f *Function) FormData() url.Values {
	return url.Values{}
}

func (a *NyckelAPIDef) ListFunction() ([]Function, error) {
	collection := []Function{}
	resp := a.getNykelResponseRead(&collection, nil, ListFunctionsEndpoint)
	if resp.RawError != nil {
		if a.panicOnError {
			panic("ListFunction: " + resp.RawError.Error())
		}
		return nil, resp.RawError
	}
	list := resp.Result.(*[]Function)
	return *list, nil
}

func (a *NyckelAPIDef) FindFunctionById(id string) (Function, error) {
	fn := Function{}
	resp := a.getNykelResponseRead(&fn, nil, fmt.Sprintf(FindFunctionsEndpoint, id))
	if resp.RawError != nil {
		if a.panicOnError {
			panic("FindFunctionById: " + resp.RawError.Error())
		}
		return Function{}, resp.RawError
	}
	return fn, nil
}

func (a *NyckelAPIDef) CreateFunction(name string, in FunctionInputType, out FunctionOutputType) (Function, error) {
	formData := url.Values{}
	formData.Add("name", name)
	formData.Add("input", in.String())
	formData.Add("output", out.String())

	result := &Function{}

	resp := a.getNykelResponseForm(formData, result, a.accessToken, CreateFunctionsEndpoint)
	if resp.RawError != nil {
		if a.panicOnError {
			panic("CreateFunction: " + resp.RawError.Error())
		}
		return Function{}, resp.RawError
	}
	f := resp.Result.(*Function)
	return *f, nil
}

//
// Image Sample
//

func (a *NyckelAPIDef) ListSamples(funcId string, count, start, end int, externalId string) ([]Sample, error) {
	collection := []Sample{}
	value := url.Values{}

	if count > 0 {
		value.Add("count", fmt.Sprint(count))
	}
	if start >= 0 {
		value.Add("start", fmt.Sprint(start))
	}
	if end >= 0 {
		value.Add("end", fmt.Sprint(end))
	}
	if externalId != "" {
		value.Add("externalId", externalId)
	}
	resp := a.getNykelResponseRead(&collection, nil, fmt.Sprintf(ListSamplesEndpoint, funcId))
	if resp.RawError != nil {
		if a.panicOnError {
			panic("ListFunction: " + resp.RawError.Error())
		}
		return nil, resp.RawError
	}
	list := resp.Result.(*[]Sample)
	return *list, nil

}

func (a *NyckelAPIDef) DeleteSample(funcId string, sampleId string) error {
	path := fmt.Sprintf(DeleteSampleEndpoint, funcId, sampleId)
	resp := a.getNyckelResponseRaw(nil, "", nil, "DELETE", a.accessToken, path)
	if resp.RawError != nil {
		if a.panicOnError {
			panic("DeleteSample: " + resp.RawError.Error())
		}
		return resp.RawError
	}
	return nil
}

func (a *NyckelAPIDef) createImageSample(funcId string, fileName string, formData map[string]string) (*ImageSample, error) {
	sample := ImageSample{}
	result := a.getNykelResponseFile(fileName, formData, &sample, a.accessToken, fmt.Sprintf(CreateImageSampleEndpoint, funcId))
	if result.RawError != nil {
		// if a.panicOnError {
		// 	panic(result.Error.Error())
		// }
		return nil, result
	}
	returnSample := result.Result.(*ImageSample)
	return returnSample, nil

}

func (a *NyckelAPIDef) CreateImageSample(funcId string, fileName string, labelName string, ourId string) (*ImageSample, error) {
	fd := map[string]string{
		"labelName": labelName,
	}
	if ourId != "" {
		fd["externalId"] = ourId
	}
	return a.createImageSample(funcId, fileName, fd)
}

func (a *NyckelAPIDef) CreateImageSampleId(funcId string, fileName string, labelId string, ourId string) (*ImageSample, error) {
	fd := map[string]string{
		"labelId": labelId,
	}
	if ourId != "" {
		fd["externalId"] = ourId
	}
	return a.createImageSample(funcId, fileName, fd)
}

// Invoke
func (a *NyckelAPIDef) InvokeFunction(funcId string, filepath string, capture bool, ourId string) (Invocation, error) {
	fd := map[string]string{
		"capture": fmt.Sprint(capture),
	}
	if ourId != "" {
		fd["externalId"] = ourId
	}
	invoc := Invocation{}
	result := a.getNykelResponseFile(filepath, fd, &invoc, a.accessToken, fmt.Sprintf(InvokeFunctionEndpoint, funcId))
	if result.RawError != nil {
		if a.panicOnError {
			panic(result.Error())
		}
		return Invocation{}, result
	}
	returnSample := result.Result.(*Invocation)
	return *returnSample, nil

}

//
// Annotate
//

func (a *NyckelAPIDef) annotateSample(funcId string, sampleId string, param map[string]any) (string, error) {
	out := map[string]any{}
	result := a.getNykelResponseJson(true, param, &out, a.accessToken, fmt.Sprintf(AnnotateSampleEndpoint, funcId, sampleId))
	if result.RawError != nil {
		if a.panicOnError {
			panic("AnnotateSample:" + result.RawError.Error())
		}
		return "", result.RawError
	}
	var labelId string
	raw, ok := out["labelId"]
	if ok {
		s, ok := raw.(string)
		if ok {
			labelId = s
		}
	}
	return labelId, nil
}

func (a *NyckelAPIDef) AnnotateSample(funcId string, sampleId string, labelName string) (string, error) {
	return a.annotateSample(funcId, sampleId, map[string]any{"labelName": labelName})
}
func (a *NyckelAPIDef) AnnotateSampleId(funcId string, sampleId string, labelId string) (string, error) {
	return a.annotateSample(funcId, sampleId, map[string]any{"labelId": labelId})
}

//
// Std API Path
//

func (a *NyckelAPIDef) getNykelResponseRead(out any, value url.Values, endpoint string) *NyckelCallResult {
	var enc string
	if value != nil {
		enc = value.Encode()
	}
	finalEndpoint := endpoint
	if enc != "" {
		finalEndpoint = endpoint + "?" + enc
	}
	return a.getNyckelResponseRaw(nil, "", out, "GET", a.accessToken, finalEndpoint)
}

func (a *NyckelAPIDef) getNykelResponseJson(isPut bool, mustEncode any, out any, auth *AccessToken, endpoint string) *NyckelCallResult {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	if err := enc.Encode(mustEncode); err != nil {
		return WrapSimpleError(err, 0)
	}
	requestBody := strings.NewReader(buf.String())
	verb := "POST"
	if isPut {
		verb = "PUT"
	}

	return a.getNyckelResponseRaw(requestBody, "application/json", out, verb, auth, endpoint)
}
func (a *NyckelAPIDef) getNykelResponseForm(formData url.Values, out any, auth *AccessToken, endpoint string) *NyckelCallResult {
	requestBody := strings.NewReader(formData.Encode())
	return a.getNyckelResponseRaw(requestBody, "application/x-www-form-urlencoded", out, "POST", auth, endpoint)
}

func (a *NyckelAPIDef) getNykelResponseFile(filename string, formData map[string]string, out any, auth *AccessToken, endpoint string) *NyckelCallResult {

	formData["filename"] = "@" + filename

	ct, body, err := createForm(formData)
	if err != nil {
		if a.panicOnError {
			panic(err)
		}
		return WrapSimpleError(err, 0)
	}

	return a.getNyckelResponseRaw(body, ct, out, "POST", auth, endpoint)
}

func (a *NyckelAPIDef) getNyckelResponseRaw(requestBody io.Reader, contentType string, out any, verb string, auth *AccessToken, endpoint string) *NyckelCallResult {

	req, err := http.NewRequest(verb, endpoint, requestBody)
	if err != nil {
		return WrapSimpleError(err, 0)
	}
	if auth != nil {
		tokenInfo := fmt.Sprintf("%s %s", auth.TokenType, auth.AccessToken)
		//log.Printf("token info %s, %v", tokenInfo[:80], req == nil)
		req.Header.Add("Authorization", tokenInfo)
	}

	if verb == "POST" || verb == "PUT" {
		req.Header.Add("Content-Type", contentType)
	}

	client := &http.Client{}
	var resp *http.Response

	//log.Printf("--->>> About to %s on %s", verb, req.URL.Path)

	resp, err = client.Do(req)
	if err != nil {
		return WrapSimpleError(err, resp.StatusCode)
	}

	// so can examine after the fact
	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return WrapSimpleError(err, resp.StatusCode)
	}
	// clip := buf.Len()
	// if buf.Len() > clipLen {
	// 	clip = clipLen
	// }
	//log.Printf("RESPONSE %d, len %d, content %s", resp.StatusCode, buf.Len(), buf.String()[:clip])
	resp.Body.Close()
	if resp.StatusCode != 200 {
		if buf.Len() > 5 && buf.String()[0:6] == "<html>" {
			return WrapSimpleError(fmt.Errorf("html error"), resp.StatusCode)
		}
		if buf.Len() > 0 {
			nerr := ErrNyckel{}
			dec := json.NewDecoder(buf)
			err := dec.Decode(&nerr)
			//helpful for debug
			if err != nil {
				return WrapSimpleError(err, resp.StatusCode)
			}
			//log.Printf("------- %s\n", buf.String())
			return &NyckelCallResult{
				StatusCode: resp.StatusCode,
				RawError:   NewErrNyckel(nerr.Text),
			}
		} else {
			return WrapSimpleError(fmt.Errorf("status code: %d", resp.StatusCode), resp.StatusCode)
		}
	}

	if buf.Len() > 0 {
		dec := json.NewDecoder(buf)
		if err := dec.Decode(out); err != nil {
			return WrapSimpleError(err, resp.StatusCode)
		}
	}
	return &NyckelCallResult{
		StatusCode: resp.StatusCode,
		RawError:   nil,
		Result:     out,
		Buffer:     buf,
	}
}

//
// ErrNykel
//

func WrapSimpleError(err error, code int) *NyckelCallResult {
	return &NyckelCallResult{
		StatusCode: code,
		RawError:   err,
	}
}

//
// Misc
//

func dumpBody(resp *http.Response) {
	fmt.Printf("\n")
	_, err := io.Copy(os.Stdout, resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n")
	resp.Body.Close()
}

//
// EnvVar
//

func envVarOrPanic(name string) string {
	cand := os.Getenv(name)
	if cand == "" {
		panic(fmt.Sprintf("could not find env var '%s'", name))
	}
	return cand
}

// helper for creating multipart file
func createForm(form map[string]string) (string, io.Reader, error) {
	body := new(bytes.Buffer)
	mp := multipart.NewWriter(body)
	defer mp.Close()
	for key, val := range form {
		if strings.HasPrefix(val, "@") {
			val = strings.TrimPrefix(val, "@")
			file, err := os.Open(val)
			if err != nil {
				return "", nil, err
			}
			defer file.Close()
			part, err := mp.CreateFormFile(key, val)
			if err != nil {
				return "", nil, err
			}
			io.Copy(part, file)
		} else {
			mp.WriteField(key, val)
		}
	}
	return mp.FormDataContentType(), body, nil
}
