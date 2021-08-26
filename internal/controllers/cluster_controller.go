package controllers

import (
	"github.com/duc-cnzj/mars/internal/response"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsapi "k8s.io/metrics/pkg/apis/metrics"
)

type ClusterController struct{}

func NewClusterController() *ClusterController {
	return &ClusterController{}
}

func (cc *ClusterController) Info(ctx *gin.Context) {
	response.Success(ctx, 200, utils.ClusterInfo())
}

var supportedMetricsAPIVersions = []string{
	"v1beta1",
}

func SupportedMetricsAPIVersionAvailable(discoveredAPIGroups *metav1.APIGroupList) bool {
	for _, discoveredAPIGroup := range discoveredAPIGroups.Groups {
		if discoveredAPIGroup.Name != metricsapi.GroupName {
			continue
		}
		for _, version := range discoveredAPIGroup.Versions {
			for _, supportedVersion := range supportedMetricsAPIVersions {
				if version.Version == supportedVersion {
					return true
				}
			}
		}
	}
	return false
}
