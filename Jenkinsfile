#!groovy

def repository = "registry.c.test-chamber-13.lan"
def repositoryCreds = "harbor-repository-creds"

def shortCommit
def workspace

def label = "kubernetes-${UUID.randomUUID().toString()}"
def templateName = "pipeline-worker"
pipeline {
    agent {
        kubernetes {
            yaml functions.podYaml(
                repo: repository,
                templateName: templateName,
                kaniko: true,
                alpine: true,
                [
                    [
                        name: "sonar",
                        path: "${repository}/library/sonarscanner:latest",
                        command: "/bin/sh"
                    ],
                    [
                        name: "golang",
                        path: "${repository}/dockerhub/library/golang:alpine",
                        command: "/bin/sh"
                    ]
                ]
            )
        }
    }

    stages {
        stage('Clone Repository') {
            steps {
                script {
                    checkout ([$class: "GitSCM",
                        branches: scm.branches,
                        extensions: scm.extensions + [$class: 'CloneOption', shallow: true],
                        userRemoteConfigs: scm.userRemoteConfigs,
                    ])
                    shortCommit = sh(returnStdout: true, script: "git log -n 1 --pretty=format:'%h'").trim()
                }
            }
        }

        stage ('Initalize Jenkins') {
            parallel {
                stage ('Set Workspace') {
                    steps {
                        script {
                            workspace = pwd()
                        }
                    }
                }

                stage ('Prepare SonarScanner') {
                    steps {
                        script {
                            def sonarScannerConfig = """
sonar.projectKey=${env.JOB_BASE_NAME.replace(" ", "-")}
sonar.projectVersion=${shortCommit}

sonar.sources=.
sonar.exclusions=**/*_test.go,**/vendor/**,**/testdata/*,html/**

sonar.tests=.
sonar.test.inclusions=**/*_test.go
sonar.test.exclusions=**/vendor/**
sonar.go.coverage.reportPaths=cover.out
"""
                            writeFile file: 'sonar-project.properties', text: sonarScannerConfig
                        }
                    }
                }
            }
        }

        stage ('Run Tests') {
            steps {
                container ('golang') {
                    script {
                        writeFile(file: workspace + "/test-chamber-13.lan.root.crt", text: functions.getCurrentRootCA())
                        writeFile(file: workspace + "/test-chamber-13.lan.ret.root.crt", text: functions.getRetiredRootCA())
                        sh """
                            ls -lah "${workspace}"
                            if [ ! "/usr/bin/curl" ] || [ ! -x "/usr/bin/curl" ]; then
                                apk add --no-cache curl
                            fi
                            if [ ! "/usr/bin/git" ] || [ ! -x "/usr/bin/git" ]; then
                                apk add --no-cache git
                                git config --global --add safe.directory '${workspace}'
                            fi
                            apk add --no-cache gcc musl-dev
                            curl \
                                --silent \
                                --location \
                                --cacert <( printf '%s\\n' "\$(cat "${workspace}/test-chamber-13.lan.root.crt")" "\$(cat "${workspace}/test-chamber-13.lan.ret.root.crt")" ) \
                                https://nexus.c.test-chamber-13.lan/repository/github-releases/jstemmer/go-junit-report/releases/download/v1.0.0/go-junit-report-v1.0.0-linux-amd64.tar.gz \
                            | tar -z -x -f - -C /usr/local/bin
                            ln -s "${workspace}" "/go/src/${env.JOB_BASE_NAME}"
                            cd "/go/src/${env.JOB_BASE_NAME}"
                            go test -short -coverprofile=cover.out `go list ./... | grep -v vendor/`
                            go test -v ./... 2>&1 | go-junit-report > report.xml
                        """
                    }
                }
            }
        }

        stage ('SonarQube Analysis') {
            steps {
                container ('sonar') {
                    script {
                        try {
                            withSonarQubeEnv('SonarQube') {
                                sh "sonar-scanner --define sonar.host.url=https://sonar.c.test-chamber-13.lan"
                            }
                        } catch(ex) {
                            unstable('Unable to communicate with Sonarqube or Sonarqube sumission failed.')
                        }
                    }
                }
            }
        }

        stage ('Build & Push') {
            steps {
                container ('kaniko') {
                    script {
                        declarativeFunctions.buildContainerMultipleDestinations(
                            dockerFile: readFile(file: "${workspace}/Dockerfile"),
                            repositoryAccess: [
                                [
                                    repository: repository,
                                    credentials: repositoryCreds
                                ],
                            ],
                            destination: [
                                "${repository}/library/webhook:latest",
                            ]
                        )
                    }
                }
            }
        }

        stage('Submit Testing Report to Jenkins') {
            steps {
                script {
                    catchError{
                        junit 'report.xml'
                    }
                }
            }
        }
    }
}
