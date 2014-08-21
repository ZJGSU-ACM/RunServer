#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <dirent.h>
#include <unistd.h>
#include <time.h>
#include <stdarg.h>
#include <ctype.h>
#include <sys/wait.h>
#include <sys/ptrace.h>
#include <sys/types.h>
#include <sys/user.h>
#include <sys/syscall.h>
#include <sys/time.h>
#include <sys/resource.h>
#include <sys/signal.h>
#include <sys/stat.h>
#include <assert.h>

#include "config.h"

int main(int argc, char** argv){
    int pid;

    int lang = atoi(argv[1]);
    char *workdir = argv[2];

    chdir(workdir);
    
    const char * CP_C[] = { "gcc", "Main.c", "-o", "Main","-Wall", "-lm",
                            "--static", "-std=c99", "-DONLINE_JUDGE", NULL
                          };
    const char * CP_X[] = { "g++", "Main.cc", "-o", "Main", "-Wall",
                            "-lm", "--static","-std=c++0x", "-DONLINE_JUDGE", NULL
                          };
	const char * CP_J[] = { "javac", "-J-Xms32m", "-J-Xmx256m", "Main.java",NULL };

    //char javac_buf[4][16];
    //char *CP_J[5];
    // for(int i=0; i<4; i++) CP_J[i]=javac_buf[i];
    // sprintf(CP_J[0],"javac");
    // sprintf(CP_J[1],"-J%s",java_xms);
    // sprintf(CP_J[2],"-J%s",java_xmx);
    // sprintf(CP_J[3],"Main.java");
    //CP_J[4]=(char *)NULL;

    pid = fork();
    if (pid == 0){
        struct rlimit LIM;
        LIM.rlim_max = 600;
        LIM.rlim_cur = 600;
        setrlimit(RLIMIT_CPU, &LIM);

        LIM.rlim_max = 900 * STD_MB;
        LIM.rlim_cur = 900 * STD_MB;
        setrlimit(RLIMIT_FSIZE, &LIM);

        LIM.rlim_max =  STD_MB<<11;
        LIM.rlim_cur =  STD_MB<<11;
        setrlimit(RLIMIT_AS, &LIM);
        
        freopen("ce.txt", "w", stdout);//record copmile error
        
        switch (lang){
        case LangC:
            execvp(CP_C[0], (char * const *) CP_C);
            break;
        case LangCC:
            execvp(CP_X[0], (char * const *) CP_X);
            break;
        case LangJava:
            execvp(CP_J[0], (char * const *) CP_J);
            break;
        default:
            exit(-1);
        }
        exit(0);
    }else{
        int status=0;
        waitpid(pid, &status, 0);
        if(status == 0){
            exit(0);
        }else{
            exit(1);
        }
    }
}