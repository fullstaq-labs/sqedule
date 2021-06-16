package dbmodels

import (
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"gorm.io/gorm"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReleaseBackgroundJob", func() {
	var db *gorm.DB
	var err error

	Describe("creation", func() {
		var org Organization
		var app Application
		var release Release

		BeforeEach(func() {
			db, err = dbutils.SetupTestDatabase()
			Expect(err).ToNot(HaveOccurred())

			err = db.Transaction(func(tx *gorm.DB) error {
				org, err = CreateMockOrganization(tx, nil)
				Expect(err).ToNot(HaveOccurred())
				app, err = CreateMockApplicationWith1Version(tx, org, nil, nil)
				Expect(err).ToNot(HaveOccurred())
				release, err = CreateMockReleaseWithInProgressState(tx, org, app, nil)
				Expect(err).ToNot(HaveOccurred())
				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		})

		It("works", func() {
			err = db.Transaction(func(tx *gorm.DB) error {
				_, numTries, err := createReleaseBackgroundJobWithDebug(tx, org.ID, app.ID, release, 1)
				Expect(err).ToNot(HaveOccurred())
				Expect(numTries).To(BeNumerically("==", 1))

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		})

		It("picks a random lock ID when the auto-incremented lock ID is already in use", func() {
			err = db.Transaction(func(tx *gorm.DB) error {
				release2, err := CreateMockReleaseWithInProgressState(tx, org, app, nil)
				Expect(err).ToNot(HaveOccurred())

				// Create a job and delete it, in order to predict what the next lock ID will be.
				job, err := CreateReleaseBackgroundJob(tx, org.ID, app.ID, release)
				Expect(err).ToNot(HaveOccurred())
				nextLockSubID := (job.LockSubID + 1) % ReleaseBackgroundJobMaxLockSubID
				err = tx.Delete(&job).Error
				Expect(err).ToNot(HaveOccurred())

				// Create a job with the predicted next lock ID.
				job = ReleaseBackgroundJob{
					BaseModel: BaseModel{
						OrganizationID: org.ID,
					},
					ApplicationID: app.ID,
					ReleaseID:     release2.ID,
					LockSubID:     nextLockSubID,
				}
				err = tx.Create(&job).Error
				Expect(err).ToNot(HaveOccurred())

				// Create another job, whose autoincremented lock ID should conflict.
				job, numTries, err := createReleaseBackgroundJobWithDebug(tx, org.ID, app.ID, release, 100)
				Expect(err).ToNot(HaveOccurred())
				Expect(numTries).To(BeNumerically("==", 2))
				Expect(job.LockSubID).ToNot(Equal((nextLockSubID+1)%ReleaseBackgroundJobMaxLockSubID),
					"Expect lock sub-ID to be random, not auto-incremented")

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		})

		It("gives up after too many lock ID picks", func() {
			err = db.Transaction(func(tx *gorm.DB) error {
				release2, err := CreateMockReleaseWithInProgressState(tx, org, app, nil)
				Expect(err).ToNot(HaveOccurred())

				// Create a job and delete it, in order to predict what the next lock ID will be.
				job, err := CreateReleaseBackgroundJob(tx, org.ID, app.ID, release)
				Expect(err).ToNot(HaveOccurred())
				nextLockSubID := (job.LockSubID + 1) % ReleaseBackgroundJobMaxLockSubID
				err = tx.Delete(&job).Error
				Expect(err).ToNot(HaveOccurred())

				// Create a job with the predicted next lock ID.
				job = ReleaseBackgroundJob{
					BaseModel: BaseModel{
						OrganizationID: org.ID,
					},
					ApplicationID: app.ID,
					ReleaseID:     release2.ID,
					LockSubID:     nextLockSubID,
				}
				err = tx.Create(&job).Error
				Expect(err).ToNot(HaveOccurred())

				// Create another job, whose autoincremented lock ID should conflict.
				_, _, err = createReleaseBackgroundJobWithDebug(tx, org.ID, app.ID, release, 1)
				Expect(err).To(MatchError("Unable to find a free lock sub-ID after 1 tries"))

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
