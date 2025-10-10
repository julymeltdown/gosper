package audio

import (
    "strings"
    "unicode"
    "gosper/internal/domain"
)

// ResolveDeviceID returns the best-matching device ID for a user-provided selector,
// which may be an exact ID, exact name (case-insensitive), prefix of the name,
// substring, or fuzzy approximate match. Returns empty string if no match.
func ResolveDeviceID(devs []domain.Device, sel string) string {
    if sel == "" { return "" }
    // 1) exact ID
    for _, d := range devs { if d.ID == sel { return d.ID } }
    // 2) exact name (case-insensitive)
    for _, d := range devs { if strings.EqualFold(d.Name, sel) { return d.ID } }

    nsel := normalize(sel)
    // 3) prefix match on normalized name
    for _, d := range devs { if strings.HasPrefix(normalize(d.Name), nsel) { return d.ID } }
    // 4) substring match
    for _, d := range devs { if strings.Contains(normalize(d.Name), nsel) { return d.ID } }
    // 5) fuzzy by normalized Levenshtein ratio
    bestID := ""
    bestScore := 0.0
    for _, d := range devs {
        score := fuzzyRatio(nsel, normalize(d.Name))
        if score > bestScore {
            bestScore = score
            bestID = d.ID
        }
    }
    if bestScore >= 0.6 { return bestID }
    return ""
}

func normalize(s string) string {
    s = strings.ToLower(s)
    var b strings.Builder
    for _, r := range s {
        if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
            b.WriteRune(r)
        }
    }
    return strings.Join(strings.Fields(b.String()), " ")
}

func fuzzyRatio(a, b string) float64 {
    if a == b { return 1 }
    if a == "" || b == "" { return 0 }
    d := levenshtein(a, b)
    maxLen := float64(len(a))
    if len(b) > len(a) { maxLen = float64(len(b)) }
    return 1 - float64(d)/maxLen
}

func levenshtein(a, b string) int {
    ra := []rune(a)
    rb := []rune(b)
    da := make([]int, len(rb)+1)
    db := make([]int, len(rb)+1)
    for j := 0; j <= len(rb); j++ { da[j] = j }
    for i := 1; i <= len(ra); i++ {
        db[0] = i
        for j := 1; j <= len(rb); j++ {
            cost := 0
            if ra[i-1] != rb[j-1] { cost = 1 }
            del := da[j] + 1
            ins := db[j-1] + 1
            sub := da[j-1] + cost
            m := del
            if ins < m { m = ins }
            if sub < m { m = sub }
            db[j] = m
        }
        da, db = db, da
    }
    return da[len(rb)]
}

