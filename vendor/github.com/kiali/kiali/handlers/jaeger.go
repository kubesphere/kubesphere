package handlers

import (
	"net/http"

	"github.com/kiali/kiali/config"
	"github.com/kiali/kiali/models"
)

// Get JaegerInfo provides the proxy Jaeger URL
func GetJaegerInfo(w http.ResponseWriter, r *http.Request) {
	jaegerConfig := config.Get().ExternalServices.Jaeger
	info := models.JaegerInfo{
		URL: jaegerConfig.URL,
	}

	// Check if URL is in the configuration
	if info.URL == "" {
		RespondWithError(w, http.StatusNotFound, "You need to set the Jaeger URL configuration.")
		return
	}

	// Check if URL is valid
	_, error := validateURL(info.URL)
	if error != nil {
		RespondWithError(w, http.StatusNotAcceptable, "You need to set a correct format for Jaeger URL in the configuration error: "+error.Error())
		return
	}

	RespondWithJSON(w, http.StatusOK, info)
}
