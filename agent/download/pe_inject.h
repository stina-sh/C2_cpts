
#ifndef PE_INJECT_H
#define PE_INJECT_H

#include <windows.h>
#include <stddef.h>  
#include <stdbool.h> 


PBYTE RemotePeExec(PBYTE pPeBuffer, LPCSTR cRemoteProcessImage, LPCSTR cProcessParms);

#endif
