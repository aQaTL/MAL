package main

import (
	"testing"

	"github.com/aqatl/mal/anilist"
)

func TestParseScore(t *testing.T) {
	{
		_, err := parseScore("0", anilist.Point10)
		if err != nil {
			t.Error(err)
		}
	}
	{
		score, err := parseScore("-1", anilist.Point10)
		if err == nil {
			t.Error("Expected fail, got", score)
		}
	}
	{
		score, err := parseScore("-1", anilist.Point10Decimal)
		if err == nil {
			t.Error("Expected fail, got", score)
		}
	}
	{
		score, err := parseScore("-1.5", anilist.Point10Decimal)
		if err == nil {
			t.Error("Expected fail, got", score)
		}
	}
	{
		score, err := parseScore("10.1", anilist.Point10Decimal)
		if err == nil {
			t.Error("Expected fail, got", score)
		}
	}
	{
		_, err := parseScore("10", anilist.Point10Decimal)
		if err != nil {
			t.Error(err)
		}
	}
	{
		_, err := parseScore("10.0", anilist.Point10Decimal)
		if err != nil {
			t.Error(err)
		}
	}
	{
		score, err := parseScore("5.50", anilist.Point10Decimal)
		if err == nil {
			t.Error("Expected fail, got", score)
		}
	}
	{
		_, err := parseScore("5.1", anilist.Point10Decimal)
		if err != nil {
			t.Error(err)
		}
	}
	{
		score, err := parseScore("5.1", anilist.Point10)
		if err == nil {
			t.Error("Expected fail, got", score)
		}
	}
	{
		score, err := parseScore("11", anilist.Point10)
		if err == nil {
			t.Error("Expected fail, got", score)
		}
	}
	{
		_, err := parseScore("10", anilist.Point10)
		if err != nil {
			t.Error(err)
		}
	}
	{
		_, err := parseScore("3", anilist.Point3)
		if err != nil {
			t.Error(err)
		}
	}
	{
		score, err := parseScore("4", anilist.Point3)
		if err == nil {
			t.Error("Expected fail, got", score)
		}
	}
	{
		_, err := parseScore("5", anilist.Point5)
		if err != nil {
			t.Error(err)
		}
	}
	{
		_, err := parseScore("5", anilist.Point5)
		if err != nil {
			t.Error(err)
		}
	}
	{
		_, err := parseScore("100", anilist.Point100)
		if err != nil {
			t.Error(err)
		}
	}
	{
		score, err := parseScore("101", anilist.Point100)
		if err == nil {
			t.Error("Expected fail, got", score)
		}
	}
}
