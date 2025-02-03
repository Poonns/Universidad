import math
a = float(input('Introduce el primer lado: '))
b = float(input('Introduce el segundo lado: '))
c = float(input('Introduce el ángulo que forman: '))

area = ((a * b) / 2) * math.sin(c)

print(f'El área del triángulo es {area}')