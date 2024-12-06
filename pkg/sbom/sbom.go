// package sbom

// import (
// 	"context"
// 	"fmt"

// 	"github.com/anchore/syft/syft"
// 	"github.com/anchore/syft/syft/format"
// 	"github.com/anchore/syft/syft/format/syftjson"
// )

// func Get(
// 	input string,
// ) ([]byte, error) {
// 	src, err := syft.GetSource(context.Background(), input, nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get input for create sbom: %w", err)
// 	}
// 	s, err := syft.CreateSBOM(context.Background(), src, nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get sbom: %s", err)
// 	}
// 	bytes, err := format.Encode(*s, syftjson.NewFormatEncoder())
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to format sbom: %s", err)
// 	}
// 	return bytes, nil
// }
