// Copyright (c) 2018 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package testdata

import (
	"errors"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/m3db/m3/src/aggregator/aggregation"
	"github.com/m3db/m3/src/aggregator/aggregator"
	maggregation "github.com/m3db/m3metrics/aggregation"
	"github.com/m3db/m3metrics/metadata"
	"github.com/m3db/m3metrics/metric"
	"github.com/m3db/m3metrics/metric/aggregated"
	"github.com/m3db/m3metrics/metric/id"
	metricid "github.com/m3db/m3metrics/metric/id"
	"github.com/m3db/m3metrics/metric/unaggregated"
	"github.com/m3db/m3metrics/pipeline/applied"
	"github.com/m3db/m3metrics/policy"
	xtime "github.com/m3db/m3x/time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
)

func test11() {
	fmt.Println("import fmt", time.Now())
	var _ = config.Configuration{}
	var _ = models.Tags{}
	var _ = httpd.Handler{}
	var _ = logging.WithContext(context.TODO())
	var _ = client.NewOptions()
	var _ = xtime.Millisecond
	var _ = kingpin.Arg("", "")
	var _ = zap.String("address", "")
	var _, _ = units.ParseBase2Bytes("test")
	var _ = template.HTMLEscapeString("test")
	var _ = quantile.NewHighBiased(1)
	var _ = parse.NewIdentifier("test")
}
