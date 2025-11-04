//go:build mipsle
// +build mipsle

TEXT ·SyscallWrite(SB), $0-24
    MOVW $2, R2 // #define SYS_write 4004
    MOVW fd+0(FP), R4
    MOVW write_buf+4(FP), R5
    MOVW nbytes+16(FP), R6
    SYSCALL
    MOVW R2, ret+0(FP)
    RET

TEXT ·SyscallHintLen(SB), $0-4
    MOVW $0xF0, R2 // #define SYS_hint_len 0xF0
    SYSCALL
    MOVW R2, ret+0(FP)
    RET

TEXT ·SyscallHintRead(SB), $0-16
    MOVW $0xF1, R2 // #define SYS_hint_read 0xF1
    MOVW ptr+0(FP), R4
    MOVW len+12(FP), R5
    SYSCALL
    RET

TEXT ·SyscallCommit(SB), $0-8
	MOVW index+0(FP), R4   // a0 = index
	MOVW word+4(FP),  R5   // a1 = word
	MOVW $0x10, R2         // v0 = syscall 4001
	SYSCALL
	RET

TEXT ·SyscallExit(SB), $0-4
	MOVW code+0(FP), R4    // a0 = code
	MOVW $0, R2         // v0 = syscall 0
	SYSCALL
	RET

TEXT ·SyscallKeccakSponge(SB), $0-8
	MOVW $0x01010009, R2   // v0 = KECCAK_SPONGE syscall
	MOVW input+0(FP), R4   // a0 = input pointer
	MOVW result+4(FP), R5  // a1 = result pointer
	SYSCALL
	RET
