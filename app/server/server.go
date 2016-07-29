package server

import (
	"path"
	"strings"

	"github.com/Unknwon/com"
	"github.com/lunny/log"
	"github.com/lunny/tango"
	"gopkg.in/inconshreveable/log15.v2"

	"io/ioutil"
	"path/filepath"

)

// Server is simple built-in http server
type Server struct {
	Tango *tango.Tango // use tango

	dstDir string
	sourDir string
	rootDir string
	prefix string
	editMd bool
}

// New returns new server
// set dstDir to make sure read correct static file
func New(dstDir string,sourDir string,rootDir string,editMd bool) *Server {
	t := tango.New([]tango.Handler{
		tango.Return(),
		tango.Param(),
		tango.Contexts(),
		tango.Recovery(true),
	}...)
	t.Logger().(*log.Logger).SetOutputLevel(log.Lerror)
	return &Server{
		Tango:  t,
		dstDir: dstDir,
		sourDir: sourDir,
		rootDir: rootDir,
		editMd: editMd,
		prefix: "/",
	}
}

// SetPrefix sets prefix to trim url
func (s *Server) SetPrefix(prefix string) {
	if prefix == "" {
		prefix = "/"
	}
	s.prefix = prefix
}

// Run starts server
func (s *Server) Run(addr string) {
	log15.Info("Server.Start." + addr)
	s.Tango.Use(logger())
	s.Tango.Get("/", s.globalHandler)

	s.Tango.Get("/*name", s.globalHandler)
	if(s.editMd){
		s.Tango.Get("/admin/*name", s.adminHandler)
		s.Tango.Get("/admin/mdlist", s.adminListMdFileHandler)
		s.Tango.Get("/admin/getmd", s.adminGetMdFileHandler)
		s.Tango.Post("/admin/updatemd", s.adminUpdateMdFileHandler)
	}
	s.Tango.Run(addr)
}

func (s *Server) serveFile(ctx *tango.Context, file string) bool {
	log15.Debug("Dest.File." + file)
	if com.IsFile(file) {
		ctx.ServeFile(file)
		return true
	}
	return false
}

func (s *Server) serveFiles(ctx *tango.Context, param string) bool {
	ext := path.Ext(param)
	if ext == "" || ext == "." {
		if !strings.HasSuffix(param, "/") {
			if s.serveFile(ctx, path.Join(s.dstDir, param+".html")) {
				return true
			}
		}
		if s.serveFile(ctx, path.Join(s.dstDir, param, "index.html")) {
			return true
		}
	}
	if s.serveFile(ctx, path.Join(s.dstDir, param)) {
		return true
	}
	return false
}

func (s *Server) globalHandler(ctx *tango.Context) {
	param := ctx.Param("*name")
	if param == "favicon.ico" || param == "robots.txt" {
		if !s.serveFiles(ctx, param) {
			ctx.NotFound()
		}
		return
	}

	if !strings.HasPrefix("/"+param, s.prefix) {
		ctx.Redirect(s.prefix)
		return
	}
	param = strings.TrimPrefix("/"+param, s.prefix)
	s.serveFiles(ctx, param)
}
func (s *Server) adminHandler(ctx *tango.Context) {

	params := ctx.Params()

	dir := s.rootDir+"/ext/admin"
	sfile := filepath.Join(dir, (*params)[0].Value)
	if len(*params) <= 0 {
		ctx.HandleError()
		return
	}



	ctx.ServeFile(sfile)
}
func (s *Server) adminListMdFileHandler(ctx *tango.Context){

	mdfiles := make([]string, 0, 10)
	dir, err := ioutil.ReadDir(s.sourDir+"/post")
	if err != nil {
		ctx.Write([]byte("{'status':2002}"))
		return
	}
	for _, fi := range dir {
		if fi.IsDir() { // 忽略目录
			continue
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), ".MD") { //匹配文件
			mdfiles = append(mdfiles, fi.Name())
		}
	}
	var result = make(map[string]interface{})
	result["data"] = mdfiles
	result["status"] = 2001
	ctx.ServeJson(result)
}
func (s *Server) adminGetMdFileHandler(ctx *tango.Context){

	md := ctx.Form("md")
	mdfilepath := s.sourDir+"/post/"+md

	content,err := ioutil.ReadFile(mdfilepath)
	var result = make(map[string]interface{})


	if(err !=nil){
		result["data"] = ""
		result["status"] = 2002
	}else{
		result["data"] = string(content)
		result["status"] = 2001
	}

	ctx.ServeJson(result)
}
func (s *Server) adminUpdateMdFileHandler(ctx *tango.Context){

	content := ctx.Form("content")
	md := ctx.Form("md")
	mdfilepath := s.sourDir+"/post/"+md

	err := ioutil.WriteFile(mdfilepath,[]byte(content),0666)

	if(err != nil){
		ctx.Write([]byte("{'status':2002}"))
	}
	ctx.Write([]byte("{'status':2001}"))

}