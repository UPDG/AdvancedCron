package main

type Config struct {
	Tasks []struct {
		Name    string
		Time    string
		Command string
		Output  *string
	}
}
