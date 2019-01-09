package main

type Config struct {
	Tasks []struct {
		Name    string
		Time    string
		User    *string
		Command string
		Output  *string
	}
}
