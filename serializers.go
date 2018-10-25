package main

import (
	j "github.com/ricardolonga/jsongo"
)

type serializerFlags struct {
	SkipNested bool
}

func (pr *pageResult) Serialize(flags serializerFlags) j.O {
	return j.Object().
		Put("url", pr.URL).
		Put("status", pr.Status)
}

func (cr *crawling) Serialize(flags serializerFlags) j.O {
	serialized := j.Object().
		Put("id", cr.ID).
		Put("url", cr.URL).
		Put("createdAt", cr.CreatedAt).
		Put("processed", cr.Processed)

	if flags.SkipNested {
		return serialized
	}

	return serialized.Put("pageResults", serializePageResults(cr.PageResults, serializerFlags{}))
}

func serializeCrawlings(crs []crawling, flags serializerFlags) *j.A {
	serialized := j.Array()

	for _, cr := range crs {
		serialized.Put(cr.Serialize(flags))
	}

	return serialized
}

func serializePageResults(prs []pageResult, flags serializerFlags) *j.A {
	serialized := j.Array()

	for _, pr := range prs {
		serialized.Put(pr.Serialize(flags))
	}

	return serialized
}
