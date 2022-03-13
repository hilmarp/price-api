package formatters

import (
	"testing"
)

func TestGetURLWithoutWWW(t *testing.T) {
	url := GetURLWithoutWWW("https://tl.is/product")
	if url != "https://tl.is/product" {
		t.Errorf("Got %s, want %s", url, "https://tl.is/product")
	}

	url = GetURLWithoutWWW("http://tl.is/product")
	if url != "http://tl.is/product" {
		t.Errorf("Got %s, want %s", url, "http://tl.is/product")
	}

	url = GetURLWithoutWWW("https://www.tl.is/product")
	if url != "https://tl.is/product" {
		t.Errorf("Got %s, want %s", url, "https://tl.is/product")
	}

	url = GetURLWithoutWWW("http://www.tl.is/product")
	if url != "http://tl.is/product" {
		t.Errorf("Got %s, want %s", url, "http://tl.is/product")
	}
}

func TestGetURLWithWWW(t *testing.T) {
	url := GetURLWithWWW("https://tl.is/product")
	if url != "https://www.tl.is/product" {
		t.Errorf("Got %s, want %s", url, "https://www.tl.is/product")
	}

	url = GetURLWithWWW("http://tl.is/product")
	if url != "http://www.tl.is/product" {
		t.Errorf("Got %s, want %s", url, "http://www.tl.is/product")
	}

	url = GetURLWithWWW("https://www.tl.is/product")
	if url != "https://www.tl.is/product" {
		t.Errorf("Got %s, want %s", url, "https://www.tl.is/product")
	}

	url = GetURLWithWWW("http://www.tl.is/product")
	if url != "http://www.tl.is/product" {
		t.Errorf("Got %s, want %s", url, "http://www.tl.is/product")
	}
}

func TestGetURLWithoutQuery(t *testing.T) {
	url := GetURLWithoutQuery("https://heimkaup.is/nuby-gomlaga-snud-glow?vid=28743")
	if url != "https://heimkaup.is/nuby-gomlaga-snud-glow" {
		t.Errorf("Got %s, want %s", url, "https://heimkaup.is/nuby-gomlaga-snud-glow")
	}

	url = GetURLWithoutQuery("https://heimkaup.is/nokia-hulstur-cc-3057-fyrir-lumia-620?vid=9881")
	if url != "https://heimkaup.is/nokia-hulstur-cc-3057-fyrir-lumia-620" {
		t.Errorf("Got %s, want %s", url, "https://heimkaup.is/nokia-hulstur-cc-3057-fyrir-lumia-620")
	}
}

func TestGetURLWithQueryParam(t *testing.T) {
	url := GetURLWithQueryParam("https://ht.is/product/uppthvottavel-60cm", "utm_source", "verdfra")
	if url != "https://ht.is/product/uppthvottavel-60cm?utm_source=verdfra" {
		t.Errorf("Got %s, want %s", url, "https://ht.is/product/uppthvottavel-60cm?utm_source=verdfra")
	}

	url = GetURLWithQueryParam("https://www.rumfatalagerinn.is/stok-vara/VILDBJERG-svefnstoll/?PathId=b701f957", "utm_source", "verdfra")
	if url != "https://www.rumfatalagerinn.is/stok-vara/VILDBJERG-svefnstoll/?PathId=b701f957&utm_source=verdfra" {
		t.Errorf("Got %s, want %s", url, "https://www.rumfatalagerinn.is/stok-vara/VILDBJERG-svefnstoll/?PathId=b701f957&utm_source=verdfra")
	}
}

func TestGetCleanURL(t *testing.T) {
	url := GetCleanURL("https://byko.is/gardurinn-og-pallurinn/gardurinn/gardahold?ProductID=166411", []string{"ProductID"})
	if url != "https://byko.is/gardurinn-og-pallurinn/gardurinn/gardahold?ProductID=166411" {
		t.Errorf("Got %s, want %s", url, "https://byko.is/gardurinn-og-pallurinn/gardurinn/gardahold?ProductID=166411")
	}

	url = GetCleanURL("https://heimkaup.is/nuby-gomlaga-snud-glow?vid=28743", []string{"ProductID"})
	if url != "https://heimkaup.is/nuby-gomlaga-snud-glow" {
		t.Errorf("Got %s, want %s", url, "https://heimkaup.is/nuby-gomlaga-snud-glow")
	}

	url = GetCleanURL("https://byko.is/gardurinn-og-pallurinn/gardurinn/gardahold?CategoryID=166411", []string{"ProductID"})
	if url != "https://byko.is/gardurinn-og-pallurinn/gardurinn/gardahold" {
		t.Errorf("Got %s, want %s", url, "https://byko.is/gardurinn-og-pallurinn/gardurinn/gardahold")
	}

	url = GetCleanURL("https://byko.is/gardurinn-og-pallurinn/gardurinn/gardahold?CategoryID=166411&ProductID=166411", []string{"ProductID"})
	if url != "https://byko.is/gardurinn-og-pallurinn/gardurinn/gardahold?ProductID=166411" {
		t.Errorf("Got %s, want %s", url, "https://byko.is/gardurinn-og-pallurinn/gardurinn/gardahold?ProductID=166411")
	}

	url = GetCleanURL("https://byko.is/gardurinn-og-pallurinn/gardurinn/gardahold?CategoryID=166411&ProductID=166411", []string{"ProductID", "CategoryID"})
	if url != "https://byko.is/gardurinn-og-pallurinn/gardurinn/gardahold?CategoryID=166411&ProductID=166411" {
		t.Errorf("Got %s, want %s", url, "https://byko.is/gardurinn-og-pallurinn/gardurinn/gardahold?CategoryID=166411&ProductID=166411")
	}

	url = GetCleanURL("https://byko.is/gardurinn-og-pallurinn/gardurinn/gardahold?CategoryID=166411&ProductID=166411", []string{})
	if url != "https://byko.is/gardurinn-og-pallurinn/gardurinn/gardahold" {
		t.Errorf("Got %s, want %s", url, "https://byko.is/gardurinn-og-pallurinn/gardurinn/gardahold")
	}
}
