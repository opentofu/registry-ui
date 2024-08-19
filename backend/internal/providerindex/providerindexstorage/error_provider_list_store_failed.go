package providerindexstorage

type ProviderListStoreFailedError struct {
	BaseError
}

func (p *ProviderListStoreFailedError) Error() string {
	if p.Cause != nil {
		return "Provider list could not be stored: " + p.Cause.Error()
	}
	return "Provider list could not be stored."
}
