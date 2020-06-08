; multiplication program
start:
 lda product
 add factor1
 sta product
 lda factor2
 sub dec
 sta factor2
 jz  exit
 jmp start

exit:
 lda product
 out
 hlt

 .org 12
dec:
 .byte 1 ; decrementer (must be 1)
product:
 .byte 0 ; result
factor1:
 .byte 2 ; first number to multiply
factor2:
 .byte 4 ; second number to multiply