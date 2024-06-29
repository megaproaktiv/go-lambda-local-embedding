package hugoembedding

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestPath2Link(t *testing.T) {
	type args struct {
		path             string
		conversionMethod int
		metadata         string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test pearls",
			args: args{
				path:             "/Users/gglawe/Documents/projects/community/2024/pearls/content/post/2023/play-ac-prompts-from-s3/index.md",
				conversionMethod: 1,
				metadata:         "2022-07-30",
			},
			want: "post/2023/play-ac-prompts-from-s3/",
		},
		{
			name: "Test pearls w/o date",
			args: args{
				path:             "/Users/gglawe/Documents/projects/community/2024/pearls/content/post/2023/play-ac-prompts-from-s3/index.md",
				conversionMethod: 1,
				metadata:         "",
			},
			want: "post/2023/play-ac-prompts-from-s3/",
		},
		{
			name: "Test aws-blog-de",
			args: args{
				path:             "/Users/gglawe/letsblog/abd/content/post/2012/amazon-aws-services-mit-beta-status.md",
				conversionMethod: 2,
				metadata:         "Wed, 12 Dec 2012 15:14:59 +0000",
			},
			want: "2012/12/amazon-aws-services-mit-beta-status.html",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Path2Link(tt.args.path, tt.args.conversionMethod, tt.args.metadata); got != tt.want {
				t.Errorf("Path2Link() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTryParseDateMonth(t *testing.T) {
	type args struct {
		dateStr string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test parse month from aws-blog.de old article",
			args: args{
				dateStr: "Wed, 12 Dec 2012 15:14:59 +0000",
			},
			want:    "12",
			wantErr: false,
		},
		{
			name: "Test parse month from aws-blog.de new article",
			args: args{
				dateStr: "2024-03-04",
			},
			want:    "03",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TryParseDateMonth(tt.args.dateStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("TryParseDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, *got, tt.want)
		})
	}
}
