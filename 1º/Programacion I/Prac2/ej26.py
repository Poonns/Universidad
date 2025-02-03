n = int(input('Introduce un numero: '))
divisores = 0
for i in range (1, n+1):
    if n % i== 0:
        divisores +=1

if divisores == 2:
    print('Es primo')
else:
    print('No es primo')