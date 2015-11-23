package monitor

import (
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/nanopack/yoke/config"
	"github.com/nanopack/yoke/state/mock"
)

func TestSingle(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	me := mock_state.NewMockState(ctrl)
	other := mock_state.NewMockState(ctrl)

	perform := start(me, other, test)
	defer perform.Stop()

	me.EXPECT().SetDBRole("single").Return(nil)

	err := perform.Single()
	if err != nil {
		test.Log(err)
		test.FailNow()
	}

}

func TestActive(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	me := mock_state.NewMockState(ctrl)
	other := mock_state.NewMockState(ctrl)

	perform := start(me, other, test)
	defer perform.Stop()

	other.EXPECT().GetDataDir().Return("test", nil)
	other.EXPECT().Location().Return("127.0.0.1:1234")

	other.EXPECT().SetSynced(true).Return(nil)

	me.EXPECT().SetDBRole("active").Return(nil)

	err := perform.Active()
	if err != nil {
		test.Log(err)
		test.FailNow()
	}

}

func TestBackup(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	me := mock_state.NewMockState(ctrl)
	other := mock_state.NewMockState(ctrl)

	perform := start(me, other, test)
	defer perform.Stop()

	me.EXPECT().HasSynced().Return(false, nil)
	me.EXPECT().HasSynced().Return(true, nil)
	me.EXPECT().SetDBRole("backup").Return(nil)

	err := perform.Backup()
	if err != nil {
		test.Log(err)
		test.FailNow()
	}

}

func start(me, other *mock_state.MockState, test *testing.T) *performer {
	// i need to create a tmp folder
	// start a action
	// then have it transition to different states.

	// overwrite the builtin config options
	config.Conf = config.Config{
		PGPort:      4567,
		DataDir:     os.TempDir() + "/postgres/",
		StatusDir:   os.TempDir() + "/postgres/",
		SyncCommand: "true",
		SystemUser:  config.SystemUser(),
	}

	perform := NewPerformer(me, other, config.Conf)

	// ignore all errors that come across this way
	go func() {
		<-perform.err
	}()

	err := perform.Initialize()
	if err != nil {
		test.Log(err)
		test.FailNow()
	}

	if err := config.ConfigureHBAConf("127.0.0.1"); err != nil {
		test.Log(err)
		test.FailNow()
	}
	if err := config.ConfigurePGConf("127.0.0.1", config.Conf.PGPort); err != nil {
		test.Log(err)
		test.FailNow()
	}

	err = perform.Start()
	if err != nil {
		test.Log(err)
		test.FailNow()
	}

	return perform
}
