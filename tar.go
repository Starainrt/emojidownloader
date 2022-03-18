package main

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"sync"
)

type TarFile struct {
	gw       *gzip.Writer
	tw       *tar.Writer
	destPath string
	fw       *os.File
	mu       sync.Mutex
}

func NewTar(dstPath string) (*TarFile, error) {
	var res TarFile
	var err error
	res.fw, err = os.Create(dstPath)
	if err != nil {
		return nil, err
	}
	res.gw = gzip.NewWriter(res.fw)
	res.tw = tar.NewWriter(res.gw)
	res.destPath = dstPath
	return &res, nil
}

func (tf *TarFile) AddFile(path, name string) error {
	tf.mu.Lock()
	defer tf.mu.Unlock()
	fso, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fso.Close()
	stats, err := fso.Stat()
	if err != nil {
		return err
	}
	hdr, err := tar.FileInfoHeader(stats, "")
	if err != nil {
		return err
	}
	hdr.Name = name
	err = tf.tw.WriteHeader(hdr)
	if err != nil {
		return err
	}
	_, err = io.Copy(tf.tw, fso)
	return err
}

func (tf *TarFile) Finish() error {
	err := tf.tw.Close()
	if err != nil {
		return err
	}
	err = tf.gw.Close()
	if err != nil {
		return err
	}
	return tf.fw.Close()
}
