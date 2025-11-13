pipeline {
    agent { label 'x86-4-32-s' }

    options {
        timestamps ()
        timeout(time: 1, unit: 'HOURS')
        disableConcurrentBuilds(abortPrevious: true)
    }

    environment {
        // Go options
        GOGC = '50'
        GOMEMLIMIT = '28GiB'

        // Aida CLI options
        STATEDB = '--db-impl carmen --db-variant go-file --carmen-schema 5'
        ARCHIVE = '--archive --archive-variant s5'
        PRIME = '--update-buffer-size 4000'
        VM = '--vm-impl lfvm'
        AIDADB = '--aida-db /mnt/substate-opera-mainnet/aida-db --substate-encoding rlp --chainid 250'
        TMPDB = '--db-tmp /mnt/tmp-disk'
        DBSRC = '/mnt/tmp-disk/state_db_carmen_go-file_${TOBLOCK}'
        PROFILE = '--cpu-profile cpu-profile.dat --memory-profile mem-profile.dat --memory-breakdown'

        // Other parameters
        TRACEDIR = 'tracefiles'
        FROMBLOCK = 'opera'
        TOBLOCK = '4570000'
    }

    stages {
        stage('Validate commit') {
            steps {
                script {
                    def CHANGE_REPO = sh (script: "basename -s .git `git config --get remote.origin.url`", returnStdout: true).trim()
                    build job: '/Utils/Validate-Git-Commit', parameters: [
                        string(name: 'Repo', value: "${CHANGE_REPO}"),
                        string(name: 'Branch', value: "${env.CHANGE_BRANCH}"),
                        string(name: 'Commit', value: "${GIT_COMMIT}")
                    ]
                }
            }
        }

        stage('Check license headers') {
            steps {
                sh 'cd scripts/license && chmod +x add_license_header.sh && ./add_license_header.sh --check'
            }
        }

        stage('Run tests') {
            stages {
                stage('Check formatting') {
                    steps {
                        catchError(buildResult: 'FAILURE', stageResult: 'FAILURE', message: 'Test Suite had a failure') {
                            sh '''diff=`find . \\( -path ./carmen -o -path ./tosca -o -path ./sonic \\) -prune -o -name '*.go' -exec gofmt -s -l {} \\;`
                                  echo $diff
                                  test -z $diff
                               '''
                        }
                    }
                }

                stage('Lint') {
                    steps {
                        sh "git submodule update --init --recursive"
                        sh "make install-dev-tools"
                        sh "make check"
                    }
                }

                stage('Build') {
                    steps {
                        script {
                            currentBuild.description = "Building on ${env.NODE_NAME}"
                        }
                        sh "git submodule update --init --recursive"
                        sh "make all"
                    }
                }

                stage('Run unit tests') {
                    environment {
                        CODECOV_TOKEN = credentials('codecov-uploader-0xsoniclabs-global')
                    }
                    steps {
                        catchError(buildResult: 'FAILURE', stageResult: 'FAILURE', message: 'Test Suite had a failure') {
                             sh 'go test ./... -coverprofile=coverage.txt -coverpkg=./...'
                             sh ('codecov upload-process -r 0xsoniclabs/aida -f ./coverage.txt -t ${CODECOV_TOKEN}')
                        }
                    }
                }

                stage('aida-vm') {
                    steps {
                        catchError(buildResult: 'FAILURE', stageResult: 'FAILURE', message: 'Test Suite had a failure') {
                            sh "build/aida-vm ${VM} ${AIDADB} --cpu-profile cpu-profile.dat --workers 32 --validate-tx ${FROMBLOCK} ${TOBLOCK}"
                        }
                        sh "rm -rf *.dat"
                    }
                }

                stage('aida-stochastic') {
                    steps {
                        catchError(buildResult: 'FAILURE', stageResult: 'FAILURE', message: 'Test Suite had a failure') {
                            sh "./build/aida-stochastic-sdb generate -output ./stats.json"
                            sh """./build/aida-stochastic-sdb replay \
                                ${STATEDB} --db-shadow-impl geth \
                                ${TMPDB}  \
                                --validate-state-hash \
                                --balance-range 1000000 \
                                --memory-breakdown \
                                --nonce-range 100000 \
                                --random-seed 4711 \
                                --log info 10 \
                                ./stats.json"""
                            sh "./build/aida-stochastic-sdb record ${AIDADB} 1 10000 --output ./stats.json"
                            sh """ ./build/aida-stochastic-sdb replay \
                                ${STATEDB} --db-shadow-impl geth \
                                --validate-state-hash \
                                --memory-breakdown \
                                --random-seed 4711 \
                                ${TMPDB} \
                                --log info 100 \
                                ./stats.json"""
                        }
                        sh "rm -rf *.dat"
                    }
                }

                stage('aida-fuzzing') {
                    steps {
                        sh "mkdir -p /mnt/tmp-disk/stats"
                        sh "rm -rf /mnt/tmp-disk/stats/*"
                        sh "curl -L -o /mnt/tmp-disk/stats/stats.tar.gz https://github.com/0xsoniclabs/aida/releases/download/testdata/stats.tar.gz"
                        sh "tar -xzf /mnt/tmp-disk/stats/stats.tar.gz -C /mnt/tmp-disk/stats"
                        catchError(buildResult: 'FAILURE', stageResult: 'FAILURE', message: 'Test Suite had a failure') {
                            sh "build/aida-stochastic-sdb replay ${STATEDB} ${TMPDB} --db-shadow-impl geth 20 /mnt/tmp-disk/stats/stats.json"
                        }
                        sh "rm -rf /mnt/tmp-disk/stats"
                    }
                }

                stage('aida-sdb record') {
                    steps {
                        sh "mkdir -p ${TRACEDIR}"
                        sh "rm -rf ${TRACEDIR}/*"
                        catchError(buildResult: 'FAILURE', stageResult: 'FAILURE', message: 'Test Suite had a failure') {
                            // use fixed ranges to control the priming time
                            sh "build/aida-sdb record --cpu-profile cpu-profile-0.dat --trace-file ${TRACEDIR}/trace-0.dat ${AIDADB} 1000 1500"
                            sh "build/aida-sdb record --cpu-profile cpu-profile-1.dat --trace-file ${TRACEDIR}/trace-1.dat ${AIDADB} 1501 2000"
                        }
                    }
                }

                stage('aida-sdb replay') {
                    steps {
                        catchError(buildResult: 'FAILURE', stageResult: 'FAILURE', message: 'Test Suite had a failure') {
                            sh "build/aida-sdb replay ${VM} ${STATEDB} ${TMPDB} ${AIDADB} ${PRIME} ${PROFILE} --shadow-db --db-shadow-impl geth --trace-file ${TRACEDIR}/trace-0.dat 1000 1500"
                            sh "build/aida-sdb replay ${VM} ${STATEDB} ${TMPDB} ${AIDADB} ${PRIME} ${PROFILE} --trace-dir ${TRACEDIR} 1000 2000"
                        }
                        sh "rm -rf ${TRACEDIR}"
                    }
                }

                stage('aida-vm-sdb live mode') {
                    steps {
                        catchError(buildResult: 'FAILURE', stageResult: 'FAILURE', message: 'Test Suite had a failure') {
                            sh "build/aida-vm-sdb substate ${VM} ${AIDADB} ${PRIME} ${TMPDB} ${STATEDB} --validate-tx --validate-state-hash --cpu-profile cpu-profile.dat --memory-profile mem-profile.dat --memory-breakdown --continue-on-failure ${FROMBLOCK} ${TOBLOCK} "
                        }
                        sh "rm -rf *.dat"
                    }
                }

                stage('aida-vm-sdb archive mode') {
                    steps {
                        catchError(buildResult: 'FAILURE', stageResult: 'FAILURE', message: 'Test Suite had a failure') {
                            sh "build/aida-vm-sdb substate ${VM} ${AIDADB} ${PRIME} ${TMPDB} ${STATEDB} ${ARCHIVE} ${PROFILE} --keep-db --validate-tx --validate-state-hash --continue-on-failure ${FROMBLOCK} ${TOBLOCK} "
                        }
                        sh "rm -rf *.dat"
                    }
                }

                stage('aida-vm-sdb archive-inquirer') {
                    steps {
                        catchError(buildResult: 'FAILURE', stageResult: 'FAILURE', message: 'Test Suite had a failure') {
                            sh "build/aida-vm-sdb substate ${VM} ${AIDADB} ${PRIME} ${TMPDB} ${STATEDB} ${ARCHIVE} ${PROFILE} --archive-query-rate 200 --validate-tx --continue-on-failure ${FROMBLOCK} ${TOBLOCK} "
                        }
                        sh "rm -rf *.dat"
                    }
                }

                stage('aida-vm-sdb db-src') {
                    steps {
                        catchError(buildResult: 'FAILURE', stageResult: 'FAILURE', message: 'Test Suite had a failure') {
                            sh "build/aida-vm-sdb substate ${VM} --db-src ${DBSRC} ${AIDADB} ${ARCHIVE} --validate-tx --cpu-profile cpu-profile.dat --memory-profile mem-profile.dat --memory-breakdown --continue-on-failure 4600001 4610000"
                        }
                        sh "rm -rf *.dat"
                    }
                }

                stage('aida-vm-sdb eth-tests') {
                    steps {
                        dir('eth-test-package') {
                            checkout scmGit(
                                userRemoteConfigs: [[url: 'https://github.com/ethereum/tests.git']],
                                // Last commit with GeneralStateTests
                                branches: [[name: '57935b91beceb43b68c772678bd5d8d53409ce34']]
                            )
                        }
                        catchError(buildResult: 'FAILURE', stageResult: 'FAILURE', message: 'Test Suite had a failure') {
                            sh """build/aida-vm-sdb ethereum-test \
                                --validate \
                                --evm-impl ethereum \
                                --vm-impl geth \
                                --db-impl geth \
                                ${TMPDB} \
                                --chainid 1337 \
                                --fork Cancun \
                                ${env.WORKSPACE}/eth-test-package/GeneralStateTests/stTransactionTest"""
                        }
                    }
                }

                stage('aida-vm-sdb tx-generator') {
                    steps {
                        catchError(buildResult: 'FAILURE', stageResult: 'FAILURE', message: 'Test Suite had a failure') {
                            sh """build/aida-vm-sdb tx-generator \
                                --db-impl carmen --db-variant go-file --carmen-schema 5 \
                                --db-tmp /mnt/tmp-disk \
                                --shadow-db --db-shadow-impl geth \
                                --tx-type all --block-length 10 --fork Cancun \
                                100"""
                        }
                    }
                }

                stage('aida-vm-adb validate-tx') {
                    steps {
                        catchError(buildResult: 'FAILURE', stageResult: 'FAILURE', message: 'Test Suite had a failure') {
                            sh "build/aida-vm-adb ${AIDADB} --db-src ${DBSRC} --cpu-profile cpu-profile.dat --validate-tx ${FROMBLOCK} ${TOBLOCK}"
                        }
                        sh "rm -rf *.dat"
                    }
                }

                stage('tear-down') {
                    steps {
                        sh "make clean"
                        sh "rm -rf *.dat ${TRACEDIR}"
                        sh "rm -rf /mnt/tmp-disk/state_db_carmen_go-file_${TOBLOCK}"
                    }
                }
            }
        }
    }
}
