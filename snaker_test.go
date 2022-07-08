package myddlmaker

import "testing"

func TestCamelToSnake(t *testing.T) {
	startsWithCommonInitialisms("ID")
	testcases := []struct {
		in   string
		want string
	}{
		{
			in:   "",
			want: "",
		},
		{
			in:   "One",
			want: "one",
		},
		{
			in:   "ONE",
			want: "o_n_e",
		},
		{
			in:   "ID",
			want: "id",
		},
		{
			in:   "i",
			want: "i",
		},
		{
			in:   "I",
			want: "i",
		},
		{
			in:   "ThisHasToBeConvertedCorrectlyID",
			want: "this_has_to_be_converted_correctly_id",
		},
		{
			in:   "ThisIDIsFine",
			want: "this_id_is_fine",
		},
		{
			in:   "ThisHTTPSConnection",
			want: "this_https_connection",
		},
		{
			in:   "HelloHTTPSConnectionID",
			want: "hello_https_connection_id",
		},
		{
			in:   "HTTPSID",
			want: "https_id",
		},
		{
			in:   "OAuthClient",
			want: "oauth_client",
		},
	}

	for _, tc := range testcases {
		got := camelToSnake(tc.in)
		if got != tc.want {
			t.Errorf("want %q, got %q", tc.want, got)
		}
	}
}

func BenchmarkCamelToSnake(b *testing.B) {
	for i := 0; i < b.N; i++ {
		camelToSnake("BenchmarkCamelToSnake")
	}
}
