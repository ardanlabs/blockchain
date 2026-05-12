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

TEXT ·SyscallHintWrite(SB), $0-16
    MOVW $0x02, R2            // v0 = WRITE (Ziren syscall code 0x02)
    MOVW $4, R4               // a0 = fd = FD_HINT = 4
    MOVW write_buf+0(FP), R5  // a1 = buf pointer (first field of slice)
    MOVW nbytes+12(FP), R6   // a2 = nbytes (after slice: ptr=0, len=4, cap=8, nbytes=12)
    SYSCALL
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

TEXT ·SyscallEnterUnconstrained(SB), $0-4
	MOVW $0x03, R2         // v0 = ENTER_UNCONSTRAINED
	SYSCALL
	MOVW R2, ret+0(FP)     // return value
	RET

TEXT ·SyscallExitUnconstrained(SB), $0-0
	MOVW $0x04, R2         // v0 = EXIT_UNCONSTRAINED
	SYSCALL
	RET

TEXT ·SyscallKeccakSponge(SB), $0-8
	MOVW $0x01010009, R2   // v0 = KECCAK_SPONGE syscall
	MOVW input+0(FP), R4   // a0 = input pointer
	MOVW result+4(FP), R5  // a1 = result pointer
	SYSCALL
	RET

// secp256k1 elliptic curve precompiles

TEXT ·SyscallSecp256k1Add(SB), $0-8
	MOVW $0x0101000A, R2   // v0 = SECP256K1_ADD
	MOVW p+0(FP), R4       // a0 = p pointer ([16]u32, x||y LE)
	MOVW q+4(FP), R5       // a1 = q pointer ([16]u32, x||y LE)
	SYSCALL
	RET

TEXT ·SyscallSecp256k1Double(SB), $0-8
	MOVW $0x0001000B, R2   // v0 = SECP256K1_DOUBLE
	MOVW p+0(FP), R4       // a0 = p pointer ([16]u32, x||y LE)
	MOVW $0, R5            // a1 = 0 (unused)
	SYSCALL
	RET

TEXT ·SyscallSecp256k1Decompress(SB), $0-8
	MOVW $0x0001000C, R2   // v0 = SECP256K1_DECOMPRESS
	MOVW point+0(FP), R4   // a0 = point pointer (64 bytes, BE)
	MOVW isOdd+4(FP), R5   // a1 = is_odd (0 or 1)
	SYSCALL
	RET

// SHA-256 precompiles

TEXT ·SyscallSha256Extend(SB), $0-4
	MOVW $0x30010005, R2   // v0 = SHA_EXTEND
	MOVW w+0(FP), R4       // a0 = w pointer ([64]u32)
	MOVW $0, R5
	SYSCALL
	RET

TEXT ·SyscallSha256Compress(SB), $0-8
	MOVW $0x01010006, R2   // v0 = SHA_COMPRESS
	MOVW w+0(FP), R4       // a0 = w pointer ([64]u32)
	MOVW state+4(FP), R5   // a1 = state pointer ([8]u32)
	SYSCALL
	RET

// BN254 elliptic curve precompiles

TEXT ·SyscallBn254Add(SB), $0-8
	MOVW $0x0101000E, R2   // v0 = BN254_ADD
	MOVW p+0(FP), R4       // a0 = p pointer ([16]u32, x||y LE)
	MOVW q+4(FP), R5       // a1 = q pointer ([16]u32, x||y LE)
	SYSCALL
	RET

TEXT ·SyscallBn254Double(SB), $0-8
	MOVW $0x0001000F, R2   // v0 = BN254_DOUBLE
	MOVW p+0(FP), R4       // a0 = p pointer ([16]u32, x||y LE)
	MOVW $0, R5
	SYSCALL
	RET

// BN254 field arithmetic precompiles

TEXT ·SyscallBn254FpAdd(SB), $0-8
	MOVW $0x01010026, R2   // v0 = BN254_FP_ADD
	MOVW a+0(FP), R4       // a0 = a pointer ([8]u32 LE), result stored here
	MOVW b+4(FP), R5       // a1 = b pointer ([8]u32 LE)
	SYSCALL
	RET

TEXT ·SyscallBn254FpSub(SB), $0-8
	MOVW $0x01010027, R2   // v0 = BN254_FP_SUB
	MOVW a+0(FP), R4       // a0 = a pointer ([8]u32 LE), result stored here
	MOVW b+4(FP), R5       // a1 = b pointer ([8]u32 LE)
	SYSCALL
	RET

TEXT ·SyscallBn254FpMul(SB), $0-8
	MOVW $0x01010028, R2   // v0 = BN254_FP_MUL
	MOVW a+0(FP), R4       // a0 = a pointer ([8]u32 LE), result stored here
	MOVW b+4(FP), R5       // a1 = b pointer ([8]u32 LE)
	SYSCALL
	RET

TEXT ·SyscallBn254Fp2Add(SB), $0-8
	MOVW $0x01010029, R2   // v0 = BN254_FP2_ADD
	MOVW a+0(FP), R4       // a0 = a pointer ([16]u32 LE), result stored here
	MOVW b+4(FP), R5       // a1 = b pointer ([16]u32 LE)
	SYSCALL
	RET

TEXT ·SyscallBn254Fp2Sub(SB), $0-8
	MOVW $0x0101002A, R2   // v0 = BN254_FP2_SUB
	MOVW a+0(FP), R4       // a0 = a pointer ([16]u32 LE), result stored here
	MOVW b+4(FP), R5       // a1 = b pointer ([16]u32 LE)
	SYSCALL
	RET

TEXT ·SyscallBn254Fp2Mul(SB), $0-8
	MOVW $0x0101002B, R2   // v0 = BN254_FP2_MUL
	MOVW a+0(FP), R4       // a0 = a pointer ([16]u32 LE), result stored here
	MOVW b+4(FP), R5       // a1 = b pointer ([16]u32 LE)
	SYSCALL
	RET

// BLS12-381 elliptic curve precompiles

TEXT ·SyscallBls12381Add(SB), $0-8
	MOVW $0x0101001E, R2   // v0 = BLS12381_ADD
	MOVW p+0(FP), R4       // a0 = p pointer ([24]u32, x||y LE)
	MOVW q+4(FP), R5       // a1 = q pointer ([24]u32, x||y LE)
	SYSCALL
	RET

TEXT ·SyscallBls12381Double(SB), $0-8
	MOVW $0x0001001F, R2   // v0 = BLS12381_DOUBLE
	MOVW p+0(FP), R4       // a0 = p pointer ([24]u32, x||y LE)
	MOVW $0, R5
	SYSCALL
	RET

// BLS12-381 field arithmetic precompiles (Fp: [12]u32 LE, Fp2: [24]u32 LE)

TEXT ·SyscallBls12381FpAdd(SB), $0-8
	MOVW $0x01010020, R2   // v0 = BLS12381_FP_ADD
	MOVW a+0(FP), R4       // a0 = a pointer ([12]u32 LE), result stored here
	MOVW b+4(FP), R5       // a1 = b pointer ([12]u32 LE)
	SYSCALL
	RET

TEXT ·SyscallBls12381FpSub(SB), $0-8
	MOVW $0x01010021, R2   // v0 = BLS12381_FP_SUB
	MOVW a+0(FP), R4       // a0 = a pointer ([12]u32 LE), result stored here
	MOVW b+4(FP), R5       // a1 = b pointer ([12]u32 LE)
	SYSCALL
	RET

TEXT ·SyscallBls12381FpMul(SB), $0-8
	MOVW $0x01010022, R2   // v0 = BLS12381_FP_MUL
	MOVW a+0(FP), R4       // a0 = a pointer ([12]u32 LE), result stored here
	MOVW b+4(FP), R5       // a1 = b pointer ([12]u32 LE)
	SYSCALL
	RET

TEXT ·SyscallBls12381Fp2Add(SB), $0-8
	MOVW $0x01010023, R2   // v0 = BLS12381_FP2_ADD
	MOVW a+0(FP), R4       // a0 = a pointer ([24]u32 LE), result stored here
	MOVW b+4(FP), R5       // a1 = b pointer ([24]u32 LE)
	SYSCALL
	RET

TEXT ·SyscallBls12381Fp2Sub(SB), $0-8
	MOVW $0x01010024, R2   // v0 = BLS12381_FP2_SUB
	MOVW a+0(FP), R4       // a0 = a pointer ([24]u32 LE), result stored here
	MOVW b+4(FP), R5       // a1 = b pointer ([24]u32 LE)
	SYSCALL
	RET

TEXT ·SyscallBls12381Fp2Mul(SB), $0-8
	MOVW $0x01010025, R2   // v0 = BLS12381_FP2_MUL
	MOVW a+0(FP), R4       // a0 = a pointer ([24]u32 LE), result stored here
	MOVW b+4(FP), R5       // a1 = b pointer ([24]u32 LE)
	SYSCALL
	RET

// secp256r1 (P-256) elliptic curve precompiles

TEXT ·SyscallSecp256r1Add(SB), $0-8
	MOVW $0x0101002C, R2   // v0 = SECP256R1_ADD
	MOVW p+0(FP), R4       // a0 = p pointer ([16]u32, x||y LE)
	MOVW q+4(FP), R5       // a1 = q pointer ([16]u32, x||y LE)
	SYSCALL
	RET

TEXT ·SyscallSecp256r1Double(SB), $0-8
	MOVW $0x0001002D, R2   // v0 = SECP256R1_DOUBLE
	MOVW p+0(FP), R4       // a0 = p pointer ([16]u32, x||y LE)
	MOVW $0, R5
	SYSCALL
	RET

TEXT ·SyscallSecp256r1Decompress(SB), $0-8
	MOVW $0x0001002E, R2   // v0 = SECP256R1_DECOMPRESS
	MOVW point+0(FP), R4   // a0 = point pointer (64 bytes)
	MOVW isOdd+4(FP), R5   // a1 = is_odd (0 or 1)
	SYSCALL
	RET

// uint256 multiplication

TEXT ·SyscallUint256Mul(SB), $0-8
	MOVW $0x0101001D, R2   // v0 = UINT256_MUL
	MOVW x+0(FP), R4       // a0 = x pointer ([8]u32 LE)
	MOVW y+4(FP), R5       // a1 = y pointer ([8]u32 LE)
	SYSCALL
	RET
