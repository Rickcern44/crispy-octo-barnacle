package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rickcern44/cassor/internal/store"
)

func lifecycleCommand() *cobra.Command {
	c := &cobra.Command{Use: "lifecycle", Short: "Record SDD-lite phases and acceptance verification"}
	var item int64
	var phase, content string
	phaseAdd := &cobra.Command{Use: "phase", RunE: func(_ *cobra.Command, _ []string) error {
		db, e := databaseForCommand()
		if e != nil {
			return e
		}
		defer db.Close()
		v, e := store.AddPhaseRecord(db, item, phase, content)
		if e != nil {
			return e
		}
		fmt.Println(v)
		return nil
	}}
	phaseAdd.Flags().Int64Var(&item, "item", 0, "feature ID")
	phaseAdd.Flags().StringVar(&phase, "phase", "", "Intake, Explore, Define, Plan, Implement, Verify, or Record")
	phaseAdd.Flags().StringVar(&content, "content", "", "compact phase output")
	_ = phaseAdd.MarkFlagRequired("item")
	_ = phaseAdd.MarkFlagRequired("phase")
	_ = phaseAdd.MarkFlagRequired("content")
	var criteriaItem, criteriaPlan int64
	var key, title, description string
	var required bool
	criteriaAdd := &cobra.Command{Use: "criterion", RunE: func(_ *cobra.Command, _ []string) error {
		if criteriaPlan == 0 && criteriaItem == 0 {
			return fmt.Errorf("either --item or --plan is required")
		}
		db, e := databaseForCommand()
		if e != nil {
			return e
		}
		defer db.Close()
		var v store.AcceptanceCriterion
		if criteriaPlan > 0 {
			v, e = store.AddAcceptanceCriterionForPlan(db, criteriaPlan, key, title, description, required)
		} else {
			v, e = store.AddAcceptanceCriterion(db, criteriaItem, title, description)
		}
		if e != nil {
			return e
		}
		fmt.Println(v)
		return nil
	}}
	criteriaAdd.Flags().Int64Var(&criteriaItem, "item", 0, "feature ID")
	criteriaAdd.Flags().Int64Var(&criteriaPlan, "plan", 0, "draft plan ID")
	criteriaAdd.Flags().StringVar(&key, "key", "", "stable criterion key")
	criteriaAdd.Flags().StringVar(&title, "title", "", "criterion")
	criteriaAdd.Flags().StringVar(&description, "description", "", "details")
	criteriaAdd.Flags().BoolVar(&required, "required", true, "criterion blocks completion when unresolved")
	_ = criteriaAdd.MarkFlagRequired("title")
	var criterionID int64
	var status, method, evidence, recordedBy, waiverReason string
	verify := &cobra.Command{Use: "verify", RunE: func(_ *cobra.Command, _ []string) error {
		db, e := databaseForCommand()
		if e != nil {
			return e
		}
		defer db.Close()
		return store.VerifyAcceptanceCriterion(db, criterionID, status, method, evidence, recordedBy, waiverReason)
	}}
	verify.Flags().Int64Var(&criterionID, "criterion", 0, "criterion ID")
	verify.Flags().StringVar(&status, "status", "", "Passed, Failed, or Waived")
	verify.Flags().StringVar(&method, "method", "manual", "manual or automated")
	verify.Flags().StringVar(&evidence, "evidence", "", "verification evidence")
	verify.Flags().StringVar(&recordedBy, "by", "", "verifier or waiver approver")
	verify.Flags().StringVar(&waiverReason, "reason", "", "required reason for a waiver")
	_ = verify.MarkFlagRequired("criterion")
	_ = verify.MarkFlagRequired("status")
	_ = verify.MarkFlagRequired("evidence")
	c.AddCommand(phaseAdd, criteriaAdd, verify)
	return c
}
