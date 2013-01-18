package hashtree

import (
	"fmt"
	"io"
	"testing"
)

type htIOTest struct {
	out string
	in  string
}

var testdata = []htIOTest{
	{"0000000000000000000000000000000000000000000000000000000000000000", ""},
	{"6100000000000000000000000000000000000000000000000000000000000000", "a"},
	{"6162000000000000000000000000000000000000000000000000000000000000", "ab"},
	{"6162630000000000000000000000000000000000000000000000000000000000", "abc"},
	{"6162636400000000000000000000000000000000000000000000000000000000", "abcd"},
	{"6162636465000000000000000000000000000000000000000000000000000000", "abcde"},
	{"6162636465660000000000000000000000000000000000000000000000000000", "abcdef"},
	{"6162636465666700000000000000000000000000000000000000000000000000", "abcdefg"},
	{"6162636465666768000000000000000000000000000000000000000000000000", "abcdefgh"},
	{"6162636465666768690000000000000000000000000000000000000000000000", "abcdefghi"},
	{"6162636465666768696a00000000000000000000000000000000000000000000", "abcdefghij"},
	{"136a3613edfd2a9070d7e54095612fcd03ddc36eaa45990830425e5401317185", "Discard medicine more than two years old."},
	{"19510cf4dcb3892e4003fff18e579df335b4fd8f3dda38be41be2a539086588f", "He who has a shady past knows that nice guys finish last."},
	{"3abd2b63c109a78cecc379abeeb62eea4bbb715f834713515d8c75c784d94a45", "I wouldn't marry him with a ten foot pole."},
	{"631d82cdd776166f4b34cfc3afc6d3b71e295dd6a66721ea37bb4b1e9e680ed6", "Free! Free!/A trip/to Mars/for 900/empty jars/Burma Shave"},
	{"577e6ebb5ba6c21494a4acc988496f8dc8c5e9a1e2ae334e52c11d5791ca0b09", "The days of the digital watch are numbered.  -Tom Stoppard"},
	{"4e6570616c207072656d69657220776f6e27742072657369676e2e0000000000", "Nepal premier won't resign."},
	{"3fc24385607fd4798dd1c6c999b414e639df3a89b4843587fb22ac99a1723a1f", "For every action there is an equal and opposite government program."},
	{"e1f513494dac385a3b718db0f99d6de13e897c40bdfcab6848a9378ba3116ed8", "His money is twice tainted: 'taint yours and 'taint mine."},
	{"c3227ffead8d14709a8fe3d48f6cae20d5da3021eab15f56473181ec4b243234", "There is no reason for any individual to have a computer in their home. -Ken Olsen, 1977"},
	{"cf571b0e96593b11a546b04df672aee0bfdefefe80979dfb870ea3f49cdd3db0", "It's a tiny change to the code and not completely disgusting. - Bob Manchek"},
	{"73697a653a2020612e6f75743a2020626164206d616769630000000000000000", "size:  a.out:  bad magic"},
	{"b5cb7b5a942cbc80314894bda5b6219dc46f76e45871aa8f8cf0bf62132b622e", "The major problem is with sendmail.  -Mark Horton"},
	{"33e7a4bf4d47595fcd913701fb76258d8d13385e6a656a91e58f269f007e0419", "Give me a rock, paper and scissors and I will move the world.  CCFestoon"},
	{"89a8b3e8ddd9718f2e98ddd5c347404f4237455c1cfc8340216930ba2e11541a", "If the enemy is within range, then so are you."},
	{"b9423d7c079b1b3a900a1024e2946316c914e9ae2bc51d15c6fc83576692d887", "It's well we cannot hear the screams/That we create in others' dreams."},
	{"fe5cbfad3b7325427988d701d56df8548c154f0288d19b91d6302160a218b1ee", "You remind me of a TV show, but that's all right: I watch it anyway."},
	{"4320697320617320706f727461626c652061732053746f6e6568656467652121", "C is as portable as Stonehedge!!"},
	{"64ad5c4b856c09cac57cd256387c2ab94e42885b65765574fcf875724b4268a9", "Even if I could be Shakespeare, I think I should still choose to be Faraday. - A. Huxley"},
	{"60d20dfb196af20886e193c102eefc631168243fcaa1fd7cb81698842feb9963", "The fugacity of a constituent in a mixture of gases at a given temperature is proportional to its mole fraction.  Lewis-Randall Rule"},
	{"cdb74d8565967a3fd605993ee30619fb2d1a499fb16433a1192ccc1fc10ff6a0", "How can you write a big system without C++?  -Paul Glick"},
}

//test data generator
func printHashTestData() {
	c := NewTree()
	for i := 0; i < len(testdata); i++ {
		g := testdata[i]
		io.WriteString(c, g.in)
		s := fmt.Sprintf("%x", c.Sum(nil))
		fmt.Printf("{\"%v\",\n\"%v\"},\n", s, g.in)
		c.Reset()
	}
}

func TestHT(t *testing.T) {
	c := NewTree()
	for i := 0; i < len(testdata); i++ {
		g := testdata[i]
		io.WriteString(c, g.in)
		s := fmt.Sprintf("%x", c.Sum(nil))
		if s != g.out {
			t.Fatalf("testdata[%d](%s) = %s want %s", i, g.in, s, g.out)
		}
		c.Reset()
	}
}
