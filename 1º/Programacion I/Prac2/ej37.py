#Programa que lee los primeros n número de la sucesion de Fibonacci

n = int(input('Introduce un número entero: '))
a = 1 
b = 1
aux = 0
print(a,b, end=', ')

for i in range(1, n-1):
    aux = a + b
    a = b
    b = aux
    print(aux, end=', ')
