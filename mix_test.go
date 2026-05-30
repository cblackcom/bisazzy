package main

import "testing"

func TestMixUrl(t *testing.T) {
	mix1 := Mix{
		// this will be sorted before turned into url taN= parameters
		Components: []MixComponent{
			{Product: "SM 77", Percent: 50},
			{Product: "SM 14", Percent: 25},
			{Product: "GL 10", Percent: 25},
		},
		FillGel: "000000",
	}
	got := mix1.RenderUrl()
	want := "https://app.bisazza.com/process.php?req=miscela&lang=en&fillgel=000000&layout=20x20&altezza=1&ta0=25;20x20_gloss/tex_GL%2010&ta1=25;20x20_smalto/tex_SM%2014&ta2=50;20x20_smalto/tex_SM%2077"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}
