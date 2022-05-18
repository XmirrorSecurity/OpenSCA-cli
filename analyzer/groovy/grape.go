package groovy

import (
	"regexp"
	"util/model"
)

// parseGroovyFile parse deps in groovy file
func parseGroovyFile(root *model.DepTree, file *model.FileInfo) {
	// repo: @GrabResolver(name='mvnRepository', root='http://central.maven.org/maven2/')
	regs := []*regexp.Regexp{
		// @Grab('org.springframework:spring-orm:3.2.5.RELEASE')
		// @Grab('org.neo4j:neo4j-cypher:2.1.4;transitive=false')
		regexp.MustCompile(``),
		// @Grab(group='org.restlet', module='org.restlet', version='1.1.6')
		// @Grab(group='org.restlet', module='org.restlet', version='1.1.6', classifier='jdk15')
		regexp.MustCompile(``),
		// Grape.grab(group:'org.slf4j', module:'slf4j-api', version:'1.7.25')
		// Grape.grab(groupId:'com.jidesoft', artifactId:'jide-oss', version:'[2.2.1,2.3)', classLoader:loader)
	}
	for _, reg := range regs {
		_ = reg
		// match := reg.FindAllStringSubmatch(string(file.Data), -1)
	}
}
