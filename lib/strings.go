package lib

func NonEmptyStringOrNil(value string) *string {
	if len(value) == 0 {
		return nil
	} else {
		return &value
	}
}
