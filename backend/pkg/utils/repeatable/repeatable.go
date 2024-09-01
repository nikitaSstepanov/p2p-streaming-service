package repeatable

import "time"

func DoWithTries(f func() error, attempts int, duration time.Duration) error {
	var err error

	for attempts > 0 {
		if err = f(); err != nil {
			time.Sleep(duration)
			attempts--
			continue
		}

		return nil
	}

	return err
}