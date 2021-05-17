package approvalrulesprocessing

import (
	"bytes"
	"fmt"
	"log"

	"github.com/fullstaq-labs/sqedule/lib/mocking"
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/releasestate"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var _ = Describe("Background processing", func() {
	var db *gorm.DB
	var org1 dbmodels.Organization
	var clock mocking.FakeClock

	BeforeEach(func() {
		var err error

		db, err = dbutils.SetupTestDatabase()
		Expect(err).ToNot(HaveOccurred())

		org1, err = dbmodels.CreateMockOrganization(db, nil)
		Expect(err).ToNot(HaveOccurred())

		clock = mocking.FakeClock{}
	})

	Describe("ProcessInBackground", func() {
		var job dbmodels.ReleaseBackgroundJob

		BeforeEach(func() {
			txerr := db.Transaction(func(tx *gorm.DB) error {
				app, err := dbmodels.CreateMockApplicationWith1Version(db, org1, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				release, err := dbmodels.CreateMockReleaseWithInProgressState(db, org1, app, nil)
				Expect(err).ToNot(HaveOccurred())

				job, err = dbmodels.CreateMockReleaseBackgroundJob(db, org1, app, release, nil)
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(txerr).ToNot(HaveOccurred())
		})

		It("processes a ReleaseBackgroundJob in the background", func() {
			err := realProcessInBackground(db, org1.ID, job, &clock, false)
			Expect(err).ToNot(HaveOccurred())

			var count int64
			tx := db.Model(dbmodels.ReleaseBackgroundJob{}).Count(&count)
			Expect(tx.Error).ToNot(HaveOccurred())
			Expect(count).To(Equal(int64(0)))

			var release dbmodels.Release
			Expect(db.First(&release).Error).ToNot(HaveOccurred())
			Expect(release.State).To(Equal(releasestate.Approved))
		})

		It("retries on error", func() {
			buffer := bytes.NewBuffer([]byte{})
			db.Logger = logger.New(log.New(buffer, "\n", log.LstdFlags), logger.Config{LogLevel: logger.Warn})

			err := realProcessInBackground(db, org1.ID, job, &clock, true)
			Expect(err).To(HaveOccurred())

			log := buffer.String()
			Expect(log).To(ContainSubstring(fmt.Sprintf("Will retry (attempt %[1]d/%[1]d)", backgroundProcessingRetryMaxAttempts)))
			Expect(log).NotTo(ContainSubstring(fmt.Sprintf("Will retry (attempt %d/)", backgroundProcessingRetryMaxAttempts+1)))
			Expect(log).To(ContainSubstring(fmt.Sprintf("Already retried %d times, so will no longer retry", backgroundProcessingRetryMaxAttempts)))
		})
	})

	Describe("ProcessAllPendingReleasesInBackground", func() {
		var org2 dbmodels.Organization

		createMockReleaseBackgroundJob := func(org dbmodels.Organization, jobLockSubID uint32) dbmodels.ReleaseBackgroundJob {
			app, err := dbmodels.CreateMockApplicationWith1Version(db, org, nil, nil)
			Expect(err).ToNot(HaveOccurred())

			release, err := dbmodels.CreateMockReleaseWithInProgressState(db, org, app, nil)
			Expect(err).ToNot(HaveOccurred())

			job, err := dbmodels.CreateMockReleaseBackgroundJob(db, org, app, release, func(job *dbmodels.ReleaseBackgroundJob) {
				job.LockSubID = jobLockSubID
			})
			Expect(err).ToNot(HaveOccurred())

			return job
		}

		BeforeEach(func() {
			txerr := db.Transaction(func(tx *gorm.DB) error {
				var err error
				org2, err = dbmodels.CreateMockOrganization(db, func(org *dbmodels.Organization) {
					org.ID = "org2"
					org.DisplayName = "Org 2"
				})
				Expect(err).ToNot(HaveOccurred())
				createMockReleaseBackgroundJob(org1, 1)
				createMockReleaseBackgroundJob(org2, 2)
				return nil
			})
			Expect(txerr).ToNot(HaveOccurred())
		})

		It("processes all ReleaseBackgroundJobs in the background", func() {
			ProcessAllPendingReleasesInBackground(db)

			Eventually(func() (int64, error) {
				var count int64
				tx := db.Model(dbmodels.ReleaseBackgroundJob{}).Count(&count)
				return count, tx.Error
			}).Should(Equal(int64(0)))

			var releases []dbmodels.Release
			Expect(db.Find(&releases).Error).ToNot(HaveOccurred())
			Expect(releases[0].State).To(Equal(releasestate.Approved))
			Expect(releases[1].State).To(Equal(releasestate.Approved))
		})
	})
})
