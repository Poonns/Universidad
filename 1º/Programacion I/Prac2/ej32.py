#Programa que escribe los factoruiales desde 1 hasta n

n = int(input('Introduce un número entero: '))
factorial = 1

for i in range(1,  n+1):
    factorial *= i
    print(f'{n}! = {factorial}')