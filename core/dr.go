package core

import (
	"fmt"
	"math/rand"
	"psim/data"
	"psim/util"
	"time"
)

// DiscountRate is the DR Influencer
// type DiscountRate struct {
// 	T1 time.Time
// 	T2 time.Time
// 	T4 time.Time
// }

// DRInfluencer is the Influencer that predicts based on DiscountRate
type DRInfluencer struct {
	cfg    *util.AppConfig
	Delta1 int
	Delta2 int
	Delta4 int
	ID     string
}

// GetAppConfig - get cfg
func (p *DRInfluencer) GetAppConfig() *util.AppConfig {
	return p.cfg
}

// SetAppConfig - set cfg
func (p *DRInfluencer) SetAppConfig(cfg *util.AppConfig) {
	p.cfg = cfg
}

// GetDelta1 - get Delta1
func (p *DRInfluencer) GetDelta1() int {
	return p.Delta1
}

// SetDelta1 - set Delta1
func (p *DRInfluencer) SetDelta1(d int) {
	p.Delta1 = d
}

// GetDelta2 - get Delta2
func (p *DRInfluencer) GetDelta2() int {
	return p.Delta2
}

// SetDelta2 - set Delta2
func (p *DRInfluencer) SetDelta2(d int) {
	p.Delta2 = d
}

// GetDelta4 - get Delta4
func (p *DRInfluencer) GetDelta4() int {
	return p.Delta4
}

// SetDelta4 - set Delta4
func (p *DRInfluencer) SetDelta4(x int) {
	p.Delta4 = x
}

// GetID - get ID
func (p *DRInfluencer) GetID() string {
	return p.ID
}

// SetID - set ID
func (p *DRInfluencer) SetID(x string) {
	p.ID = x
}

// Init - initializes a DRInfluencer
func (p *DRInfluencer) Init(cfg *util.AppConfig, delta4 int) {
	p.cfg = cfg
	p.Delta1 = rand.Intn(30) - 30
	p.Delta2 = -1
	p.Delta4 = 3
	// fmt.Printf("*** NOTICE ***  Influencer: DR - T1, T2, and T4 are set to hardcoded test values\n")
	// p.Delta1 = util.RandomInRange(cfg.MinDelta1, cfg.MaxDelta1)
	// p.Delta2 = util.RandomInRange(cfg.MinDelta2, cfg.MaxDelta2)
	// p.Delta4 = delta4
	p.ID = fmt.Sprintf("DRInfluencer|%d|%d|%d|%s", p.Delta1, p.Delta2, p.Delta4, util.GenerateRefNo())
}

// ProfileString - a quick description of the type of Influencer and
//
//	its key attributes.
//
// ----------------------------------------------------------------------------
func (p *DRInfluencer) ProfileString() string {
	return p.ID + fmt.Sprintf(" (Discount Rate, T1 = %d, T2 = %d, T4 = %d)", p.Delta1, p.Delta2, p.Delta4)
}

// GetPrediction - using the supplied date, it researches data and makes
// a prediction on whther to "buy" or "hold"
//
// RETURNS
//
//	action     -  "buy" or "hold"
//	prediction - probability of correctness - most valid for "buy" action
//	error      - nil on success, error encountered otherwise
//
// ---------------------------------------------------------------------------
func (p *DRInfluencer) GetPrediction(t3 time.Time) (string, float64, error) {
	t1 := t3.AddDate(0, 0, p.Delta1)
	t2 := t3.AddDate(0, 0, p.Delta2)
	//---------------------------------------------------------------------------
	// Determine dDRR = (DiscountRateRatio at t1) - (DiscountRateRatio at t2)
	//---------------------------------------------------------------------------
	rec1 := data.DRFindRecord(t1)
	if rec1 == nil {
		err := fmt.Errorf("ExchangeRate Record for %s not found", t1.Format("1/2/2006"))
		return "hold", 0, err
	}
	rec2 := data.DRFindRecord(t2)
	if rec2 == nil {
		err := fmt.Errorf("ExchangeRate Record for %s not found", t2.Format("1/2/2006"))
		return "hold", 0, err
	}
	dDRR := rec1.USJPDRRatio - rec2.USJPDRRatio

	//-------------------------------------------------------------------------------
	// Prediction formula (based on the change in DiscountRateRatios):
	//     dDRR > 0:   buy on t3, sell on t4
	//     dDRR <= 0:  take no action
	//-------------------------------------------------------------------------------
	prediction := "hold"
	if dDRR > 0 {
		prediction = "buy"
		// util.DPrintf("dt1 = %s, dt2 = %s, ratio1 = %6.3f, ratio2 = %6.3f, dDRR = %6.3f, prediction: %s\n",
		// 	t1.Format("1/2/2006"), t2.Format("1/2/2006"), rec1.USJPDRRatio, rec2.USJPDRRatio, dDRR, prediction)
	}

	// todo - return proper probability
	return prediction, 0.5, nil
}
