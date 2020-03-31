package dag

import "context"

// ListVersion return server api version.
func ListVersion(ctx context.Context) (string, string, error) {
	f := mustExtractFactory(ctx)
	v, err := f.Client().DialOrDie().Discovery().ServerVersion()
	if err != nil {
		return "", "", err
	}

	return v.Major, v.Minor, nil
}
