n = int(input('Introduce un numero: '))
for i in range (1, n+1):
    if i == n:
        print(i, end ='')
    else:
        print(i, end =',')
