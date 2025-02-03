import math
rad = float(input('Introduce el radio: '))

area = math.pi * rad ** 2
longui = 2 * math.pi * rad

print(f'El área del círculo es {area:10.2f}, mientras que su longuitud es {longui:10.2f}')