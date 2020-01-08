package api

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestLists(t *testing.T) {
	checkClient(t)

	web := NewSP(spClient).Web()
	newListTitle := uuid.New().String()

	t.Run("Modifiers", func(t *testing.T) {
		lists := web.Lists()
		mods := lists.Select("*").Expand("*").Filter("*").Top(1).OrderBy("*", true).modifiers
		if mods == nil || len(mods.mods) != 5 {
			t.Error("can't add modifiers")
		}
	})

	t.Run("Get", func(t *testing.T) {
		data, err := web.Lists().Select("Id,Title").Conf(headers.verbose).Get()
		if err != nil {
			t.Error(err)
		}
		if len(data.Data()) == 0 {
			t.Error("can't get webs")
		}
		if bytes.Compare(data, data.Normalized()) == -1 {
			t.Error("wrong response normalization")
		}
	})

	t.Run("Add", func(t *testing.T) {
		if _, err := web.Lists().Add(newListTitle, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		listInfo, err := getAnyList()
		if err != nil {
			t.Error(err)
		}
		if _, err := web.Lists().GetByID(listInfo.ID).Get(); err != nil {
			t.Error(err)
		}
	})

	t.Run("GetByTitle", func(t *testing.T) {
		listInfo, err := getAnyList()
		if err != nil {
			t.Error(err)
		}
		if _, err := web.Lists().GetByTitle(listInfo.Title).Get(); err != nil {
			t.Error(err)
		}
	})

	t.Run("GetListByURI", func(t *testing.T) {
		listURI := getRelativeURL(spClient.AuthCnfg.GetSiteURL()) +
			"/Lists/" + strings.Replace(newListTitle, "-", "", -1)
		if _, err := web.GetList(listURI).Get(); err != nil {
			t.Error(err)
		}
	})

	t.Run("AddWithURI", func(t *testing.T) {
		listTitle := uuid.New().String()
		listURI := uuid.New().String()
		if _, err := web.Lists().AddWithURI(listTitle, listURI, nil); err != nil {
			t.Error(err)
		}
		if err := web.Lists().GetByTitle(listTitle).Delete(); err != nil {
			t.Error(err)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := web.Lists().GetByTitle(newListTitle).Delete(); err != nil {
			t.Error(err)
		}
	})

}

func getAnyList() (*ListInfo, error) {
	web := NewSP(spClient).Web()
	data, err := web.Lists().Select("Id,Title").Top(1).Conf(headers.verbose).Get()
	if err != nil {
		return nil, err
	}
	if len(data.Data()) == 0 {
		return nil, fmt.Errorf("can't get webs")
	}
	return data.Data()[0].Data(), nil
}
