package request

import "net/http"

// IsInternal returns true if the request is internal.
func IsInternal(r *http.Request) bool {
	return isKubernetesRequest(r)
}

// isKubernetesRequest returns true if the request is from kubernetes.
func isKubernetesRequest(r *http.Request) bool {
	// Kubernetes sets the X-Forwarded-For header when coming from the ingress, therefore we can check if the header is
	// set to determine if the request is from kubernetes.
	//
	// See: https://stackoverflow.com/questions/70164677/determine-if-http-request-to-a-service-is-within-or-outside-of-the-kubernetes-cl
	// The header will only be set from external traffic, so we can check if the header is set to determine if the
	// request is internal.

	h := r.Header.Get("X-Forwarded-For")
	// If the header is not set, then the request is internal.
	return h == ""
}
