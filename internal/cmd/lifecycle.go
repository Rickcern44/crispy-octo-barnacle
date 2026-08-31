package cmd

import (
	"fmt"
	"github.com/rickcern44/cassor/internal/store"
	"github.com/spf13/cobra"
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
	var criteriaItem int64
	var title, description string
	criteriaAdd := &cobra.Command{Use: "criterion", RunE: func(_ *cobra.Command, _ []string) error {
		db, e := databaseForCommand()
		if e != nil {
			return e
		}
		defer db.Close()
		v, e := store.AddAcceptanceCriterion(db, criteriaItem, title, description)
		if e != nil {
			return e
		}
		fmt.Println(v)
		return nil
	}}
	criteriaAdd.Flags().Int64Var(&criteriaItem, "item", 0, "feature ID")
	criteriaAdd.Flags().StringVar(&title, "title", "", "criterion")
	criteriaAdd.Flags().StringVar(&description, "description", "", "details")
	_ = criteriaAdd.MarkFlagRequired("item")
	_ = criteriaAdd.MarkFlagRequired("title")
	var criterionID int64
	var status, method, evidence string
	verify := &cobra.Command{Use: "verify", RunE: func(_ *cobra.Command, _ []string) error {
		db, e := databaseForCommand()
		if e != nil {
			return e
		}
		defer db.Close()
		return store.VerifyAcceptanceCriterion(db, criterionID, status, method, evidence)
	}}
	verify.Flags().Int64Var(&criterionID, "criterion", 0, "criterion ID")
	verify.Flags().StringVar(&status, "status", "", "Passed, Failed, or Waived")
	verify.Flags().StringVar(&method, "method", "manual", "manual or automated")
	verify.Flags().StringVar(&evidence, "evidence", "", "verification evidence")
	_ = verify.MarkFlagRequired("criterion")
	_ = verify.MarkFlagRequired("status")
	_ = verify.MarkFlagRequired("evidence")
	c.AddCommand(phaseAdd, criteriaAdd, verify)
	return c
}
