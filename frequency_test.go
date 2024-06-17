package slogrotate

import (
	"reflect"
	"testing"
	"time"
)

func TestFrequency_IsSame(t *testing.T) {
	mustParse := func(value string) time.Time {
		location, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.Local)
		if err != nil {
			panic(err)
		}
		return location
	}

	type args struct {
		lastTime    time.Time
		currentTime time.Time
	}
	tests := []struct {
		name string
		f    Frequency
		args args
		want bool
	}{
		{
			name: "minute: same",
			f:    FrequencyMinutely,
			args: args{mustParse("2006-01-02 15:04:05"), mustParse("2006-01-02 15:04:06")},
			want: true,
		},
		{
			name: "minute: not same",
			f:    FrequencyMinutely,
			args: args{mustParse("2006-01-02 15:04:05"), mustParse("2006-01-02 15:05:05")},
			want: false,
		},
		{
			name: "minute: not same 2",
			f:    FrequencyMinutely,
			args: args{mustParse("2006-01-02 15:04:05"), mustParse("2006-01-02 16:04:05")},
			want: false,
		},
		{
			name: "hour: same",
			f:    FrequencyHourly,
			args: args{mustParse("2006-01-02 15:04:05"), mustParse("2006-01-02 15:05:06")},
			want: true,
		},
		{
			name: "hour: not same",
			f:    FrequencyHourly,
			args: args{mustParse("2006-01-02 15:04:05"), mustParse("2006-01-02 16:04:05")},
			want: false,
		},
		{
			name: "hour: not same 2",
			f:    FrequencyHourly,
			args: args{mustParse("2006-01-02 15:04:05"), mustParse("2006-01-03 15:04:05")},
			want: false,
		},
		{
			name: "day: same",
			f:    FrequencyDaily,
			args: args{mustParse("2006-01-02 15:04:05"), mustParse("2006-01-02 16:05:06")},
			want: true,
		},
		{
			name: "day: not same",
			f:    FrequencyDaily,
			args: args{mustParse("2006-01-02 15:04:05"), mustParse("2006-01-03 15:04:05")},
			want: false,
		},
		{
			name: "day: not same 2",
			f:    FrequencyDaily,
			args: args{mustParse("2006-01-02 15:04:05"), mustParse("2006-02-02 15:04:05")},
			want: false,
		},
		{
			name: "weekly: same",
			f:    FrequencyWeekly,
			args: args{mustParse("2006-01-02 15:04:05"), mustParse("2006-01-08 16:05:06")},
			want: true,
		},
		{
			name: "weekly: not same",
			f:    FrequencyWeekly,
			args: args{mustParse("2006-01-02 15:04:05"), mustParse("2006-01-09 15:04:05")},
			want: false,
		},
		{
			name: "weekly: not same 2",
			f:    FrequencyWeekly,
			args: args{mustParse("2006-01-02 15:04:05"), mustParse("2007-02-02 15:04:05")},
			want: false,
		},
		{
			name: "month: same",
			f:    FrequencyMonthly,
			args: args{mustParse("2006-01-02 15:04:05"), mustParse("2006-01-03 16:05:06")},
			want: true,
		},
		{
			name: "month: not same",
			f:    FrequencyMonthly,
			args: args{mustParse("2006-01-02 15:04:05"), mustParse("2006-02-02 15:04:05")},
			want: false,
		},
		{
			name: "month: not same 2",
			f:    FrequencyMonthly,
			args: args{mustParse("2006-01-02 15:04:05"), mustParse("2007-01-02 15:04:05")},
			want: false,
		},
		{
			name: "year: same",
			f:    FrequencyYearly,
			args: args{mustParse("2006-01-02 15:04:05"), mustParse("2006-02-03 16:05:06")},
			want: true,
		},
		{
			name: "year: not same",
			f:    FrequencyYearly,
			args: args{mustParse("2006-01-02 15:04:05"), mustParse("2007-01-02 15:04:05")},
			want: false,
		},
		{
			name: "none: same",
			f:    FrequencyNone,
			args: args{mustParse("2006-01-02 15:04:05"), mustParse("2007-01-02 15:04:05")},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.IsSame(tt.args.lastTime, tt.args.currentTime); got != tt.want {
				t.Errorf("IsSame() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFrequency_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		f       Frequency
		want    []byte
		wantErr bool
	}{
		{
			name:    "ok",
			f:       FrequencyDaily,
			want:    []byte(`"daily"`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.f.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalJSON() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFrequency_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		f       Frequency
		args    args
		wantErr bool
	}{
		{
			name:    "ok",
			f:       FrequencyDaily,
			args:    args{[]byte(`"daily"`)},
			wantErr: false,
		},
		{
			name:    "err",
			f:       FrequencyDaily,
			args:    args{[]byte(`"xxx"`)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.f.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
