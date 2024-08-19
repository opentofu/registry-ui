package providerindexstorage

type ProviderListReadFailedError struct {
	BaseError
}

func (p *ProviderListReadFailedError) Error() string {
	if p.Cause != nil {
		return "Provider list could not be read: " + p.Cause.Error()
	}
	return "Provider list could not be read."
}
