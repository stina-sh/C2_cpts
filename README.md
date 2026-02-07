# C2_cpts
a minimal C2 implementation for taking the CPTS exam 

we often hear about the latest shiny new C2 framework weather it is covnant mythic or havoc and are impressed with the mind dazling amount of features they provide, 
but the reality is that this tools aren't worth the bits they'r stored on, unless the user is intimately fimiliar with the source code of the tool, enough 
to debug modify improve or customize every and each line of it 

and that's why while preparing for the cpts when i was faced with the choice between swallwing my pride to use a prebuilt framework or go through it only
using netcat and doing everything manually like a caveman :D , i decided to just write my own custom c2 from scratch (;

-


this c2 features :
-

server writen in go:
-
- handling listeners / sessions
- basic frontend for managment


-

agent writen in C and go :
-
C :  for tools injection using process hollowing [base address relocation]

go : for uploading / beaconing

-


needless to say this is still in early stages but it serve as malluable/modular C2 skeleton to improve upon 
-
-
-
-
-
-
-
⚠️ educational-only / no-liability This project is created for educational purposes .
The author is not responsible for any misuse of this software. Do not use this 
against systems you do not have explicit permission to test.
