import math
a = float(input('Introduce un valor a : '))
b = float(input('Introduce un valor b : '))
c = float(input('Introduce un valor c : '))

discriminante = (b ** 2) - (4 * a * c)

x1 = (-b + math.sqrt(discriminante)) / (2 * a)
x2 = (-b - math.sqrt(discriminante)) / (2 * a)

print(f'Las soluciones son {x1} y {x2}')