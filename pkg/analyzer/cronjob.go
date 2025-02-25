/*
Copyright 2023 The K8sGPT Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package analyzer

import (
	"fmt"
	"time"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	cron "github.com/robfig/cron/v3"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CronJobAnalyzer struct{}

func (analyzer CronJobAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	kind := "CronJob"

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	var results []common.Result

	cronJobList, err := a.Client.GetClient().BatchV1().CronJobs(a.Namespace).List(a.Context, v1.ListOptions{})
	if err != nil {
		return results, err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, cronJob := range cronJobList.Items {
		var failures []common.Failure
		if cronJob.Spec.Suspend != nil && *cronJob.Spec.Suspend {
			failures = append(failures, common.Failure{
				Text: fmt.Sprintf("CronJob %s is suspended", cronJob.Name),
				Sensitive: []common.Sensitive{
					{
						Unmasked: cronJob.Namespace,
						Masked:   util.MaskString(cronJob.Namespace),
					},
					{
						Unmasked: cronJob.Name,
						Masked:   util.MaskString(cronJob.Name),
					},
				},
			})
		} else {
			// check the schedule format
			if _, err := CheckCronScheduleIsValid(cronJob.Spec.Schedule); err != nil {
				failures = append(failures, common.Failure{
					Text: fmt.Sprintf("CronJob %s has an invalid schedule: %s", cronJob.Name, err.Error()),
					Sensitive: []common.Sensitive{
						{
							Unmasked: cronJob.Namespace,
							Masked:   util.MaskString(cronJob.Namespace),
						},
						{
							Unmasked: cronJob.Name,
							Masked:   util.MaskString(cronJob.Name),
						},
					},
				})
			}

			// check the starting deadline
			if cronJob.Spec.StartingDeadlineSeconds != nil {
				deadline := time.Duration(*cronJob.Spec.StartingDeadlineSeconds) * time.Second
				if deadline < 0 {

					failures = append(failures, common.Failure{
						Text: fmt.Sprintf("CronJob %s has a negative starting deadline", cronJob.Name),
						Sensitive: []common.Sensitive{
							{
								Unmasked: cronJob.Namespace,
								Masked:   util.MaskString(cronJob.Namespace),
							},
							{
								Unmasked: cronJob.Name,
								Masked:   util.MaskString(cronJob.Name),
							},
						},
					})

				}
			}

		}

		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", cronJob.Namespace, cronJob.Name)] = common.PreAnalysis{
				FailureDetails: failures,
			}
			AnalyzerErrorsMetric.WithLabelValues(kind, cronJob.Name, cronJob.Namespace).Set(float64(len(failures)))

		}

		for key, value := range preAnalysis {
			currentAnalysis := common.Result{
				Kind:  kind,
				Name:  key,
				Error: value.FailureDetails,
			}
			a.Results = append(results, currentAnalysis)
		}
	}

	return a.Results, nil
}

// Check CRON schedule format
func CheckCronScheduleIsValid(schedule string) (bool, error) {
	_, err := cron.ParseStandard(schedule)
	if err != nil {
		return false, err
	}

	return true, nil
}
