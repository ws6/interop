package interop

import (
	"fmt"
	"math"
	"sort"
)

//BoxWhiskerStat
//Subtile Box-whisker stat
type BoxWhiskerStat struct {
	Q1           float64
	Q2           float64 //median
	Q3           float64
	Q4           float64
	Whisker_top  float64   `json:"whisker_top"`
	Whisker_low  float64   `json:"whisker_low"`
	Whisker_high float64   `json:"whisker_high"`
	Outliers     []float64 `json:"-"`
	Mean         float64
	Stdev        float64 //standard derivation
	IQR          float64
}

func MeanStat(a *[]float64) (mean float64, stdev float64) {
	sum, devsum := float64(0.0), float64(0.0)
	arr := *a
	num := len(arr)
	if num == 0 {
		return
	}
	for _, v := range arr {
		sum += v
	}
	mean = sum / float64(num)
	for _, v := range arr {
		b := mean - v
		devsum += (b * b)
	}
	stdev = math.Sqrt(devsum / float64(num))
	return
}

func Median(numbers []float64) float64 {
	middle := len(numbers) / 2
	result := numbers[middle]
	if len(numbers)%2 == 0 {
		result = (result + numbers[middle-1]) / 2
	}
	return result
}

func (self *BoxWhiskerStat) GetFloat64(arr *[]float64) (_e error) {

	if arr == nil {
		return fmt.Errorf("array is nil")
	}
	if len(*arr) == 0 {
		return nil
	}
	numbers := *arr
	sz := len(numbers)
	if sz == 0 {
		return nil
	}
	sort.Float64s(numbers)
	//get box whisker plot from float64 array
	self.Mean, self.Stdev = MeanStat(arr)
	self.Q4 = self.Mean
	self.Q2 = Median(numbers)
	//	self.Whisker_top = self.

	middle := sz / 2
	self.Q1 = numbers[0]
	if middle > 0 && middle < sz {

		self.Q1 = Median(numbers[0:middle])
	}
	self.Q3 = numbers[0]
	if (middle+1) > 0 && (middle+1) < sz {

		self.Q3 = Median(numbers[middle+1 : sz])
	}
	IQR := self.Q3 - self.Q1
	self.IQR = IQR
	self.Whisker_high = self.Q3 + IQR
	self.Whisker_low = self.Q1 - IQR
	for _, f := range numbers {
		if f < (self.Whisker_low-1.5*IQR) || f > (self.Whisker_high+1.5*IQR) {
			self.Outliers = append(self.Outliers, f)
		}
	}
	return nil
}
func (self *BoxWhiskerStat) GetFloat32(arr []float32) error {
	f64 := []float64{}
	for _, v := range arr {
		f64 = append(f64, float64(v))
	}
	//get box whisker plot from float64 array
	return self.GetFloat64(&f64)
}
