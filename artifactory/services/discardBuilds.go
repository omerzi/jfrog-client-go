package services

import (
	"encoding/json"
	buildinfo "github.com/jfrog/build-info-go/entities"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/http/jfroghttpclient"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type DiscardBuildsService struct {
	client     *jfroghttpclient.JfrogHttpClient
	ArtDetails auth.ServiceDetails
}

func NewDiscardBuildsService(client *jfroghttpclient.JfrogHttpClient) *DiscardBuildsService {
	return &DiscardBuildsService{client: client}
}

func (ds *DiscardBuildsService) DiscardBuilds(params DiscardBuildsParams) error {
	log.Info("Discarding builds...")

	discardUrl := ds.ArtDetails.GetUrl()
	restApi := path.Join("api/build/retention/", params.GetBuildName())
	requestFullUrl, err := utils.BuildArtifactoryUrl(discardUrl, restApi, make(map[string]string))
	if err != nil {
		return err
	}
	requestFullUrl += "?async=" + strconv.FormatBool(params.IsAsync())
	if params.ProjectKey != "" {
		requestFullUrl += "&project=" + params.ProjectKey
	}

	var excludeBuilds []string
	if params.GetExcludeBuilds() != "" {
		excludeBuilds = strings.Split(params.GetExcludeBuilds(), ",")
	}

	minimumBuildDateString := ""
	if params.GetMaxDays() != "" {
		minimumBuildDateString, err = calculateMinimumBuildDate(time.Now(), params.GetMaxDays())
		if err != nil {
			return err
		}
	}

	data := DiscardBuildsBody{
		ExcludeBuilds:    excludeBuilds,
		MinimumBuildDate: minimumBuildDateString,
		MaxBuilds:        params.GetMaxBuilds(),
		DeleteArtifacts:  params.IsDeleteArtifacts()}
	requestContent, err := json.Marshal(data)
	if err != nil {
		return errorutils.CheckError(err)
	}

	httpClientsDetails := ds.getArtifactoryDetails().CreateHttpClientDetails()
	utils.SetContentType("application/json", &httpClientsDetails.Headers)

	resp, body, err := ds.client.SendPost(requestFullUrl, requestContent, &httpClientsDetails)
	if err != nil {
		return err
	}
	if err = errorutils.CheckResponseStatusWithBody(resp, body, http.StatusNoContent); err != nil {
		return err
	}
	if params.IsAsync() {
		log.Info("Builds are being discarded asynchronously.")
		return nil
	}

	log.Info("Builds discarded.")
	return nil
}

func calculateMinimumBuildDate(startingDate time.Time, maxDaysString string) (string, error) {
	maxDays, err := strconv.Atoi(maxDaysString)
	if err != nil {
		return "", errorutils.CheckError(err)
	}

	minimumBuildDate := startingDate.Add(-24 * time.Hour * (time.Duration(maxDays)))
	minimumBuildDateString := minimumBuildDate.Format(buildinfo.TimeFormat)

	return minimumBuildDateString, nil
}

func (ds *DiscardBuildsService) getArtifactoryDetails() auth.ServiceDetails {
	return ds.ArtDetails
}

type DiscardBuildsBody struct {
	MinimumBuildDate string   `json:"minimumBuildDate,omitempty"`
	MaxBuilds        string   `json:"count,omitempty"`
	ExcludeBuilds    []string `json:"buildNumbersNotToBeDiscarded,omitempty"`
	DeleteArtifacts  bool     `json:"deleteBuildArtifacts"`
}

type DiscardBuildsParams struct {
	DeleteArtifacts bool
	BuildName       string
	ProjectKey      string
	MaxDays         string
	MaxBuilds       string
	ExcludeBuilds   string
	Async           bool
}

func (bd *DiscardBuildsParams) GetBuildName() string {
	return bd.BuildName
}

func (bd *DiscardBuildsParams) GetProjectKey() string {
	return bd.ProjectKey
}

func (bd *DiscardBuildsParams) GetMaxDays() string {
	return bd.MaxDays
}

func (bd *DiscardBuildsParams) GetMaxBuilds() string {
	return bd.MaxBuilds
}

func (bd *DiscardBuildsParams) GetExcludeBuilds() string {
	return bd.ExcludeBuilds
}

func (bd *DiscardBuildsParams) IsDeleteArtifacts() bool {
	return bd.DeleteArtifacts
}

func (bd *DiscardBuildsParams) IsAsync() bool {
	return bd.Async
}

func NewDiscardBuildsParams() DiscardBuildsParams {
	return DiscardBuildsParams{}
}
