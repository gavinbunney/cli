package main

import (
  "errors"
  "strconv"
  "encoding/json"
  "github.com/parnurzeal/gorequest"
)

/*
 * {
 *   "guid": "3fbc7381a664",
 *   "namespace": "devel",
 *   "service_type": "heydiddlyho-http",
 *   "version": {
 *     "major": 0,
 *     "minor": 33
 *   }
 * }
 */
type UnitSummary struct {
  Guid string `json:"guid"`
  NamespaceRef string `json:"namespace"`
  ServiceType string `json:"service_type"`
  Version FeatureVersion `json:"version"`
}

type FeatureVersion struct {
  Major int `json:"major"`
  Minor int `json:"minor"`
}

/*
 * {
 *   "policies": [ "foo", "bar", "baz" ]
 * }
 */
type PolicyList struct {
  Policies []string `json:"policies"`
}

/////////////////// LIST ///////////////////

func ListUnits(delimitedDcs string, delimitedNamespaces string, delimitedStatuses string, http *gorequest.SuperAgent, cfg *Config) (list []UnitSummary, err []error){
  uri := "/v1/units?"
  // set the datacenters if specified
  if (isValidCommaDelimitedList(delimitedDcs)){
    uri = uri+"dc="+delimitedDcs+"&"
  }
  if (isValidCommaDelimitedList(delimitedStatuses)){
    uri = uri+"status="+delimitedStatuses+"&"
  } else {
    // if the user didnt specify statuses, they probally only want ready units.
    uri = uri+"status=ready,warming,manual,deprecated&"
  }
  if (isValidCommaDelimitedList(delimitedNamespaces)){
    uri = uri+"ns="+delimitedNamespaces
  } else {
    uri = uri+"ns=devel,qa,prod"
  }

  r, bytes, errs := AugmentRequest(
    http.Get(cfg.Endpoint+uri), cfg).EndBytes()

  if (r.StatusCode / 100 != 2){
    errs = append(errs, errors.New("bad response from Nelson server"))
    return nil, errs
  } else {
    var list []UnitSummary
    if err := json.Unmarshal(bytes, &list); err != nil {
      panic(err)
    }
    return list, errs
  }
}

func PrintListUnits(units []UnitSummary){
  var tabulized = [][]string {}
  for _,u := range units {
    tabulized = append(tabulized,[]string{ u.Guid, u.NamespaceRef, u.ServiceType, strconv.Itoa(u.Version.Major)+"."+strconv.Itoa(u.Version.Minor) })
  }

  RenderTableToStdout([]string{ "GUID", "Namespace", "Unit", "Version" }, tabulized)
}

/////////////////// DEPRECATION ///////////////////

/*
 * {
 *   "service_type": "heydiddlyho-http",
 *   "version":{
 *     "major":1,
 *     "minor":33
 *   }
 * }
 */
type DeprecationRequest struct {
  ServiceType string `json:"service_type"`
  Version FeatureVersion `json:"version"`
}

func Deprecate(req DeprecationRequest, http *gorequest.SuperAgent, cfg *Config) (str string, err []error){
  r, body, errs := AugmentRequest(
    http.Post(cfg.Endpoint+"/v1/units/deprecate"), cfg).Send(req).EndBytes()

  if (r.StatusCode / 100 != 2){
    resp := string(body[:])
    errs = append(errs, errors.New("Unexpected response from Nelson server"))
    return resp, errs
  } else {
    return "Requested deprecation of "+req.ServiceType+" "+strconv.Itoa(req.Version.Major)+"."+strconv.Itoa(req.Version.Minor), errs
  }
}
