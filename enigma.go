package main

import (
    "bufio"
    "errors"
    "flag"
    "fmt"
    "io"
    "os"
    "strings"
    "unicode"
)

type Rotor struct {
    name        string
    wiring      string
    position    int
    notch       int
    ringSetting int
}

type Enigma struct {
    rotors    [3]Rotor
    reflector string
    plugboard map[rune]rune
}

var rotorWirings = map[string]string{
    "I":    "EKMFLGDQVZNTOWYHXUSPAIBRCJ",
    "II":   "AJDKSIRUXBLHWTMCQGZNPYFVOE",
    "III":  "BDFHJLCPRTXVZNYEIWGAKMUSQO",
    "IV":   "ESOVPZJAYQUIRHXLNFTGKDCMWB",
    "V":    "VZBRGITYUPSDNHLXAWMJQOFECK",
}

var rotorNotches = map[string]int{
    "I":    17, // Q
    "II":   5,  // E
    "III":  22, // V
    "IV":   10, // J
    "V":    26, // Z
}

var reflectors = map[string]string{
    "A": "EJMZALYXVBWFCRQUONTSPIKHGD",
    "B": "YRUHQSLDPXNGOKMIEBFZCWVJAT",
    "C": "FVPJIAOYEDRZXWGCTKUQSBNMHL",
}

func NewEnigma(rotorNames []string, reflectorName string) (*Enigma, error) {
    if len(rotorNames) != 3 {
        return nil, errors.New("exactly three rotors must be specified")
    }

    reflectorWiring, exists := reflectors[reflectorName]
    if !exists {
        return nil, fmt.Errorf("invalid reflector type: %s", reflectorName)
    }

    var rotors [3]Rotor
    for i, name := range rotorNames {
        wiring, exists := rotorWirings[name]
        if !exists {
            return nil, fmt.Errorf("invalid rotor type: %s", name)
        }
        rotors[i] = Rotor{
            name:        name,
            wiring:      wiring,
            position:    0,
            notch:      rotorNotches[name],
            ringSetting: 0,
        }
    }

    return &Enigma{
        rotors:    rotors,
        reflector: reflectorWiring,
        plugboard: make(map[rune]rune),
    }, nil
}

func (e *Enigma) Reset() {
    for i := range e.rotors {
        e.rotors[i].position = 0
    }
}

func (e *Enigma) SetRotorPositions(pos1, pos2, pos3 int) error {
    if pos1 < 1 || pos1 > 26 || pos2 < 1 || pos2 > 26 || pos3 < 1 || pos3 > 26 {
        return errors.New("rotor positions must be between 1 and 26")
    }
    e.rotors[0].position = pos1 - 1
    e.rotors[1].position = pos2 - 1
    e.rotors[2].position = pos3 - 1
    return nil
}

func (e *Enigma) SetRingSettings(r1, r2, r3 int) error {
    if r1 < 1 || r1 > 26 || r2 < 1 || r2 > 26 || r3 < 1 || r3 > 26 {
        return errors.New("ring settings must be between 1 and 26")
    }
    e.rotors[0].ringSetting = r1 - 1
    e.rotors[1].ringSetting = r2 - 1
    e.rotors[2].ringSetting = r3 - 1
    return nil
}

func (e *Enigma) AddPlugboardConnection(a, b rune) error {
    a = unicode.ToUpper(a)
    b = unicode.ToUpper(b)
    
    if a < 'A' || a > 'Z' || b < 'A' || b > 'Z' {
        return errors.New("plugboard connections must be between A and Z")
    }
    
    if _, exists := e.plugboard[a]; exists {
        return fmt.Errorf("letter %c is already connected", a)
    }
    if _, exists := e.plugboard[b]; exists {
        return fmt.Errorf("letter %c is already connected", b)
    }
    
    e.plugboard[a] = b
    e.plugboard[b] = a
    return nil
}

func (e *Enigma) rotateRotors() {
    if e.rotors[1].position == (e.rotors[1].notch - 1) {
        e.rotors[1].position = (e.rotors[1].position + 1) % 26
        e.rotors[0].position = (e.rotors[0].position + 1) % 26
    } else if e.rotors[2].position == (e.rotors[2].notch - 1) {
        e.rotors[1].position = (e.rotors[1].position + 1) % 26
    }
    
    e.rotors[2].position = (e.rotors[2].position + 1) % 26
}

func (e *Enigma) encryptChar(c rune) rune {
    if c < 'A' || c > 'Z' {
        return c
    }

    e.rotateRotors()

    if val, ok := e.plugboard[c]; ok {
        c = val
    }

    pos := int(c - 'A')
    for i := 2; i >= 0; i-- {
        offset := (e.rotors[i].position - e.rotors[i].ringSetting + 26) % 26
        pos = (pos + offset) % 26
        newPos := int(e.rotors[i].wiring[pos] - 'A')
        pos = (newPos - offset + 26) % 26
    }

    pos = int(e.reflector[pos] - 'A')

    for i := 0; i < 3; i++ {
        offset := (e.rotors[i].position - e.rotors[i].ringSetting + 26) % 26
        pos = (pos + offset) % 26
        for j := 0; j < 26; j++ {
            if int(e.rotors[i].wiring[j]-'A') == pos {
                pos = j
                break
            }
        }
        pos = (pos - offset + 26) % 26
    }

    result := rune(pos + 'A')
    if val, ok := e.plugboard[result]; ok {
        result = val
    }

    return result
}

func processIO(enigma *Enigma, reader io.Reader, writer io.Writer) error {
    scanner := bufio.NewScanner(reader)
    for scanner.Scan() {
        input := scanner.Text()
        output := ""
        for _, c := range strings.ToUpper(input) {
            output += string(enigma.encryptChar(c))
        }
        if _, err := fmt.Fprintln(writer, output); err != nil {
            return fmt.Errorf("error writing output: %v", err)
        }
    }
    return scanner.Err()
}

func main() {
    rotors := flag.String("rotors", "I,II,III", "Rotor selection (e.g., I,II,III)")
    reflectorType := flag.String("reflector", "B", "Reflector type (A, B, or C)")
    pos1 := flag.Int("r1", 1, "Position of first rotor (1-26)")
    pos2 := flag.Int("r2", 1, "Position of second rotor (1-26)")
    pos3 := flag.Int("r3", 1, "Position of third rotor (1-26)")
    ring1 := flag.Int("ring1", 1, "Ring setting of first rotor (1-26)")
    ring2 := flag.Int("ring2", 1, "Ring setting of second rotor (1-26)")
    ring3 := flag.Int("ring3", 1, "Ring setting of third rotor (1-26)")
    plugboard := flag.String("p", "", "Plugboard connections (e.g., AB CD EF)")
    flag.Parse()

    rotorNames := strings.Split(*rotors, ",")
    enigma, err := NewEnigma(rotorNames, *reflectorType)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    enigma.Reset()

    if err := enigma.SetRotorPositions(*pos1, *pos2, *pos3); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    if err := enigma.SetRingSettings(*ring1, *ring2, *ring3); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    if *plugboard != "" {
        pairs := strings.Fields(*plugboard)
        for _, pair := range pairs {
            if len(pair) == 2 {
                if err := enigma.AddPlugboardConnection(rune(pair[0]), rune(pair[1])); err != nil {
                    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
                    os.Exit(1)
                }
            } else {
                fmt.Fprintf(os.Stderr, "Error: Invalid plugboard pair: %s\n", pair)
                os.Exit(1)
            }
        }
    }

    if err := processIO(enigma, os.Stdin, os.Stdout); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
