#Programa que muestra el numero factorial igual o mas grande al numero n introducido

n = int(input('Introduce un número entero: '))
factorial = 1
i = 1

while factorial <= n:
    i += 1
    factorial *= i

print(f'El número buscado es {i} ({i-1}! <= {n} < {i}!')