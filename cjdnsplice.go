/*@flow*/
// vim: sts=4:sw=4:et:ts=4
/*
 * You may redistribute this program and/or modify it under the terms of
 * the GNU General Public License as published by the Free Software Foundation,
 * either version 3 of the License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */
package cjdnsplice

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func CharToBits(b byte) []int {
	i, _ := strconv.ParseInt(string([]byte{b}), 16, 8)
	return []int{
		int((i >> 3) % 2),
		int((i >> 2) % 2),
		int((i >> 1) % 2),
		int(i % 2),
	}
}

func DecodeLabel(label string) []int {
	var out []int
	label = strings.Replace(label, ".", "", -1)
	for i := 0; i < len(label); i++ {
		out = append(out, CharToBits(label[i])...)
	}
	return out
}

func EncodeLabel(label []int) string {
	var res string
	for _, n := range label {
		if n == 0 {
			res += "0"
		} else if n == 1 {
			res += "1"
		} else {
			panic(fmt.Sprintf("invalid label %v", label))
		}
	}
	return res
}

func LabelToBits(label string) []int {
	return DecodeLabel(label)
}

func bitsToChar(bits []int) string {
	var num int64
	for i := 0; i < 4; i++ {
		num |= int64(bits[len(bits)-1-i] << uint(i))
	}
	return strconv.FormatInt(num, 16)
}

func BitsToLabel(bits []int, includeDots bool) string {
	var chars []byte
	for i := 0; i < 16; i++ {
		if includeDots && (i%4) == 0 {
			chars = append([]byte{'.'}, chars...)
		}
		chars = append([]byte(bitsToChar(bits)), chars...)
		bits = bits[0 : len(bits)-4]
	}
	if includeDots {
		chars = chars[0 : len(chars)-1]
	}
	return string(chars)
}

func errorArray() []int {
	var res [64]int
	for i := range res {
		res[i] = 1
	}
	return res[:]
}

func spliceBits1(goHere, viaHere []int) []int {
	for viaHere[0] == 0 {
		viaHere = viaHere[1:len(viaHere)]
	}
	viaHere = viaHere[1:len(viaHere)]
	goHere = append(goHere, viaHere...)
	for goHere[0] == 0 {
		goHere = goHere[1:len(goHere)]
	}

	// FIXME: these two can cancel themselves probably
	goHere = goHere[1:len(goHere)]
	goHere = append([]int{1}, goHere...)

	if len(goHere) >= 60 {
		return errorArray()
	}
	for len(goHere) < 64 {
		goHere = append([]int{0}, goHere...)
	}
	return goHere
}

func SpliceBits(arguments ...[]int) []int {
	var result []int
	if len(arguments) < 2 {
		panic("must have at least 2 args")
	}
	for i, arg := range arguments {
		if i == 0 {
			result = arg
		} else {
			result = spliceBits1(result, arg)
		}
	}
	return result
}

type EncodingForm struct {
	BitCount  int    `json:"bitCount"`
	Prefix    string `json:"prefix"`
	PrefixLen int    `json:"prefixLen"`
}

type EncodingScheme []EncodingForm

var (
	Scheme_f4  = EncodingScheme{EncodingForm{BitCount: 4, Prefix: "", PrefixLen: 0}}
	Scheme_f8  = EncodingScheme{EncodingForm{BitCount: 8, Prefix: "", PrefixLen: 0}}
	Scheme_v48 = EncodingScheme{
		EncodingForm{BitCount: 4, Prefix: "01", PrefixLen: 1},
		EncodingForm{BitCount: 8, Prefix: "00", PrefixLen: 1},
	}
	Scheme_v358 = EncodingScheme{
		EncodingForm{BitCount: 3, Prefix: "01", PrefixLen: 1},
		EncodingForm{BitCount: 5, Prefix: "02", PrefixLen: 2},
		EncodingForm{BitCount: 8, Prefix: "00", PrefixLen: 2},
	}
	Scheme_v37 = EncodingScheme{
		EncodingForm{BitCount: 3, Prefix: "01", PrefixLen: 1},
		EncodingForm{BitCount: 7, Prefix: "00", PrefixLen: 1},
	}
	Schemes = map[string]EncodingScheme{
		"f4":   Scheme_f4,
		"f8":   Scheme_f8,
		"v48":  Scheme_v48,
		"v358": Scheme_v358,
		"v37":  Scheme_v37,
	}
)

func PrintScheme(scheme EncodingScheme) string {
	schemeStr, _ := json.Marshal(scheme)
	for name, sch := range Schemes {
		str, _ := json.Marshal(sch)
		if string(str) == string(schemeStr) {
			return name
		}
	}
	return string(schemeStr)
}

func Splice(labels ...string) string {
	var bits [][]int
	for _, l := range labels {
		bits = append(bits, LabelToBits(l))
	}
	return BitsToLabel(SpliceBits(bits...), true)
}

func GetEncodingForm(lstr string, scheme EncodingScheme) int {
	for i, s := range scheme {
		var pfxStr = s.Prefix[len(s.Prefix)-s.PrefixLen : len(s.Prefix)]
		if strings.HasSuffix(lstr, pfxStr) {
			return i
		}
	}
	return -1
}

func fixLength(bits []int, length int) []int {
	for len(bits) > length {
		if bits[0] != 0 {
			panic("length cannot be reduced")
		}
		bits = bits[1:len(bits)]
	}
	for len(bits) < length {
		bits = append([]int{0}, bits...)
	}
	return bits
}

const FormCanonical = -5000

func reEncode0(labelStr string, scheme EncodingScheme, desiredFormNum int) (string, error) {
	formN := GetEncodingForm(labelStr, scheme)
	if formN < 0 {
		return "", fmt.Errorf("could not detect encoding form")
	}

	label := DecodeLabel(labelStr)
	form := scheme[formN]

	//let dir = label.splice(-(form.bitCount + form.prefixLen));
	//dir.splice(dir.length - form.prefixLen);
	dir := label[len(label)-form.BitCount-form.PrefixLen : len(label)]
	label = label[0 : len(label)-form.BitCount-form.PrefixLen]
	dir = dir[0 : len(dir)-form.PrefixLen]

	var desiredFormN = desiredFormNum
	if desiredFormNum == FormCanonical {
		desiredFormN = -1
		bitLen := len(dir)
		for _, n := range dir {
			if n == 0 {
				bitLen--
			} else {
				break
			}
		}
		for i, s := range scheme {
			if s.BitCount < bitLen {
				continue
			}
			if desiredFormN > -1 && s.BitCount >= scheme[desiredFormN].BitCount {
				continue
			}
			desiredFormN = i
		}
	}

	if desiredFormN >= len(scheme) || desiredFormN < 0 {
		return "", fmt.Errorf("invalid desired form num")
	}
	desiredForm := scheme[desiredFormN]

	if PrintScheme(scheme) == "v358" {
		// Special magic for SCHEME_358 legacy.
		if desiredFormN == 0 && EncodeLabel(dir) == "00111" || EncodeLabel(dir) == "00000111" {
			// This is a special case where encodingForm 2 looks usable but it is not
			// because number 0001 is reserved for the self-route.
			desiredFormN = 1
		}
		if formN == 0 || desiredFormN == 0 {
			dirN, _ := strconv.ParseUint(BitsToLabel(dir, false), 16, 64)
			if formN == 0 {
				switch dirN {
				case 0:
					return "", fmt.Errorf("cannot re-encode self-route")
				case 1:
					dirN = 0
				default:
					dirN--
				}
			}
			if desiredFormN == 0 {
				switch dirN {
				case 0:
					dirN = 1
				default:
					dirN++
				}
			}
			dir = nil
			for _, char := range strconv.FormatUint(dirN, 2) {
				switch char {
				case '0':
					dir = append(dir, 0)
				case '1':
					dir = append(dir, 0)
				}
			}
		}
	}

	dir = fixLength(dir, desiredForm.BitCount)
	label = append(label, dir...)
	x := LabelToBits(desiredForm.Prefix)
	label = append(label, x[len(x)-desiredForm.PrefixLen:len(x)]...)
	label = fixLength(label, 60)
	label = fixLength(label, 64)
	out := BitsToLabel(label, true)

	return out, nil
}

func ReEncode(labelStr string, scheme EncodingScheme, desiredFormNum int) (string, error) {
	res, err := reEncode0(labelStr, scheme, desiredFormNum)
	if err != nil {
		err = fmt.Errorf("Failed to reencode %s to form %d in scheme %v", labelStr, desiredFormNum, PrintScheme(scheme))
	}
	return res, err
}

/*

const isOneHop = module.exports.isOneHop =
    (label /*:string* / , encodingScheme /*:EncodingScheme* /) => {
    const formNum = getEncodingForm(label, encodingScheme);
    if (formNum < 0) { throw new Error("not a valid label for the given scheme"); }
    const form = encodingScheme[formNum];
    const bits = form.bitCount + form.prefixLen;
    const labelBits = _labelToBits(label);
    for (let i = 0; i < labelBits.length - bits - 1; i++) {
        if (labelBits[i]) { return false; }
    }
    return true;
};

/*::
type PathHop = { labelP?: string, key?: string, encodingScheme: EncodingScheme, labelN?: string };
* /
const buildLabel = module.exports.buildLabel = (pathArray /*:Array<PathHop>* /) => {
    const path = [];
    pathArray.forEach(function (hop, i) {
        let labelN = hop.labelN;
        if (!labelN) {
            if (i < pathArray.length - 1) { throw new Error("every hop must have labelN"); }
            return;
        }
        if (hop.labelP) {
            const formP = getEncodingForm(hop.labelP, hop.encodingScheme);
            const formN = getEncodingForm(labelN, hop.encodingScheme);
            const bitsP = hop.encodingScheme[formP].bitCount + hop.encodingScheme[formP].prefixLen;
            const bitsN = hop.encodingScheme[formN].bitCount + hop.encodingScheme[formN].prefixLen;
            if (bitsP > bitsN) {
                //console.log(JSON.stringify(pathArray));
                labelN = reEncode(labelN, hop.encodingScheme, formP);
            }
        }
        path.push(labelN);
    });
    //console.log(JSON.stringify(path));
    let result = path[0];
    if (path.length > 1) {
        result = splice.apply(null, path.slice(0).reverse());
    }
    return { label: result, path: path };
};
*/
