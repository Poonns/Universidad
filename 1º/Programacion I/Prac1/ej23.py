import math
a = float(input('Introduce el primer lado: '))
b = float(input('Introduce el segundo lado: '))
c = float(input('Introduce el ángulo que forman: '))

radianes = (c * math.pi) / 180
area = ((a * b) / 2) * math.sin(radianes)

print(f'El área del triángulo es {area:10.2f}')