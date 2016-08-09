package vizzini_test

import (
	"time"

	"code.cloudfoundry.org/bbs/models"
	. "code.cloudfoundry.org/vizzini/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Term", func() {
	var task *models.TaskDefinition

	BeforeEach(func() {
		task = Task()
	})

	Describe("{DOCKER} Creating a Docker-based Task", func() {
		BeforeEach(func() {
			task.RootFs = "docker:///onsi/grace-busybox"
			task.Action = models.WrapAction(&models.RunAction{
				Path: "/grace",
				Args: []string{"-catchTerminate"},
				User: "root",
			})
			task.ResultFile = "/home/alice/payload"

			Expect(bbsClient.DesireTask(logger, guid, domain, task)).To(Succeed())
			Eventually(TaskGetter(logger, guid)).Should(HaveTaskState(models.Task_Running))
		})

		It("should cancel", func() {
			Expect(bbsClient.CancelTask(logger, guid)).To(Succeed())

			By("marking the task as completed")
			task, err := bbsClient.TaskByGuid(logger, guid)
			Expect(err).NotTo(HaveOccurred())
			Expect(task.State).To(Equal(models.Task_Completed))
			Expect(task.Failed).To(BeTrue())
			Expect(task.FailureReason).To(Equal("task was cancelled"))
		})

		It("should not be able to resolve", func() {
			//bbsClient.ResolvingTask(logger, guid)
			// Expect(bbsClient.CancelTask(logger, guid)).To(Succeed())
			//Expect(models.ConvertError(err).Type).To(Equal(models.Error_InvalidStateTransition))
		})

		It("should be able to cancel then resolv and delete", func() {
			Expect(bbsClient.CancelTask(logger, guid)).To(Succeed())
			Expect(bbsClient.ResolvingTask(logger, guid)).To(Succeed())
			Expect(bbsClient.DeleteTask(logger, guid)).To(Succeed())
			_, err := bbsClient.TaskByGuid(logger, guid)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Cancelling Tasks and Starting Task", func() {
		var guids []string
		BeforeEach(func() {
			task.RootFs = "docker:///onsi/grace-busybox"
			task.Action = models.WrapAction(&models.RunAction{
				Path: "/grace",
				Args: []string{"-catchTerminate"},
				User: "root",
			})
			task.ResultFile = "/home/alice/payload"

			guids = []string{"vizzini-guid-1",
				"vizzini-guid-2",
				"vizzini-guid-3",
				"vizzini-guid-4",
				"vizzini-guid-5",
				"vizzini-guid-6",
				"vizzini-guid-7",
			}

			for i := 0; i < 7; i++ {
				Expect(bbsClient.DesireTask(logger, guids[i], domain, task)).To(Succeed())
				Eventually(TaskGetter(logger, guids[i])).Should(HaveTaskState(models.Task_Running))
			}
			time.Sleep(30 * time.Second)
		})

		AfterEach(func() {
			for i := 0; i < 7; i++ {
				Expect(bbsClient.ResolvingTask(logger, guids[i])).To(Succeed())
				Expect(bbsClient.DeleteTask(logger, guids[i])).To(Succeed())
			}
			Expect(bbsClient.ResolvingTask(logger, "failing-guid")).To(Succeed())
			Expect(bbsClient.DeleteTask(logger, "failing-guid")).To(Succeed())
		})

		FIt("should not be able to DesireTask until they have stopped", func() {
			for i := 0; i < 7; i++ {
				Expect(bbsClient.CancelTask(logger, guids[i])).To(Succeed())
			}
			Expect(bbsClient.DesireTask(logger, "failing-guid", domain, task)).To(Succeed())
			Consistently(TaskGetter(logger, "failing-guid"), 5).ShouldNot(HaveTaskState(models.Task_Running))
		})
	})

	It("should not be able to DesireTask until they have stopped", func() {
		Expect(bbsClient.CancelTask(logger, "vizzini-1-0ce7879c")).To(Succeed())
	})
})
