package cmd

import "fmt"

func All() error {
	if err := Summary(); err != nil {
		return err
	}

	fmt.Println()
	if err := Projects(); err != nil {
		return err
	}

	fmt.Println()
	if err := Sessions(nil); err != nil {
		return err
	}

	fmt.Println()
	if err := Tokens(); err != nil {
		return err
	}

	return nil
}
