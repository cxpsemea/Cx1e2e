def exit_code = 0

pipeline {
    agent any

    stages {
        stage('run') {
            steps {
                // uncomment the following line if you are setting up your own jenkinsfile
                // git branch: 'main', url: 'https://github.com/cxpsemea/cx1e2e'
                
                sh 'chmod +x get-bin.sh' 
                sh './get-bin.sh' // script will download the latest binary and set +x
                sh 'ls -al'
                // you must have a cx1e2e_admin OIDC Client set up in your CheckmarxOne tenant
                withCredentials([usernamePassword(credentialsId: "cx1e2e_admin", usernameVariable: 'OIDC_USR', passwordVariable: 'OIDC_PSW')]) {
                    script {   
                        int code = sh( script:"export E2E_RUN_SUFFIX='_'\$(date +%Y%m%d) && ./cx1e2e-bin --config ./examples/all.yaml --client \"$OIDC_USR\" --secret \"$OIDC_PSW\" --cx1 \"https://eu.ast.checkmarx.net\" --iam \"https://eu.iam.checkmarx.net\" --tenant \"tenant\"", returnStatus: true)
                        echo "Pipeline returned: ${code} tests failed"
                        exit_code = code
                        if ( code > 0 ) {
                            currentBuild.result = 'UNSTABLE'
                        }
                    }
                }
            }
        }
    }
    
    post{
        always {
            archiveArtifacts artifacts: 'cx1e2e_result.*', fingerprint: true
        }  
        // uncomment the following and set an email address - tested to work with the Email Extended extension for jenkins
        /*unstable {
            emailext to: "an email address",
                subject: "JaaS - E2E All - Failure (${exit_code} tests failed)",
                body: "The Jenkins All end-to-end test had the following failures:\n\n" + '${BUILD_LOG_REGEX, regex="^FAIL.*"}',
                attachLog: true
        } */     
    }
}
