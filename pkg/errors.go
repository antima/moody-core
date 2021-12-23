package pkg

type MoodyError string

func (err MoodyError) Error() string {
	return string(err)
}
