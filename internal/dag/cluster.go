package dag

import "context"

// ListVersion return server api version.
func ListVersion(ctx context.Context) (string, string, error) {
	f := mustExtractFactory(ctx)
	dial, err := f.Client().Dial()
	if err != nil {
		return "", "", err
	}
	v, err := dial.Discovery().ServerVersion()
	if err != nil {
		return "", "", err
	}

	return v.Major, v.Minor, nil
}
