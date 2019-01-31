package main

import (
	"reflect"
	"strings"
	"testing"
)

func Test_toCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		want    command
		wantErr bool
	}{
		{
			name: "login works",
			args: "login -u root -p calvin 10.0.0.5",
			want: command{
				Name: "login",
				Arguments: map[string][]string{
					"u": []string{"root"},
					"p": []string{"calvin"},
					"":  []string{"10.0.0.5"},
				},
			},
		},

		{
			name: "boot settings",
			args: "boot_settings -once local_cd",
			want: command{
				Name: "boot_settings",
				Arguments: map[string][]string{
					"once": []string{"true"},
					"":     []string{"local_cd"},
				},
			},
		},

		{
			name: "query multi value",
			args: "query pwState kvmEnabled voltages temperatures",
			want: command{
				Name: "query",
				Arguments: map[string][]string{
					"": []string{"pwState", "kvmEnabled", "voltages", "temperatures"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := strings.Split(tt.args, " ")
			got, err := toCommand(args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("toCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}
