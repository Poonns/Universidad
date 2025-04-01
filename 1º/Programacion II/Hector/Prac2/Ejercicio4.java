package Hector.Prac2;

import java.util.Locale;

public class Ejercicio4 {

    public static class Punto {
        private double x, y;

        public Punto(double x, double y) {
            this.x = x;
            this.y = y;
        }

        public double getX() { return x; }
        public double getY() { return y; }

        public double distancia(Punto otroPunto) {
            return Math.sqrt(Math.pow(otroPunto.getX() - this.x, 2) + Math.pow(otroPunto.getY() - this.y, 2));
        }

        @Override
        public String toString() {
            return String.format(Locale.US, "(%.2f, %.2f)", x, y);
        }
    }

    public static class Restaurante {
        private String nombre;
        private Punto posicion;
        private int valoracion;

        public Restaurante(String nombre, Punto posicion, int valoracion) {
            this.nombre = nombre;
            this.posicion = posicion;
            this.valoracion = valoracion;
        }

        public String getNombre() { return nombre; }
        public Punto getPosicion() { return posicion; }
        public int getValoracion() { return valoracion; }
    }

    public static String restauranteMejorValorado(Restaurante[] v, Punto p, double d) {
        Restaurante mejor = null;
        
        for (Restaurante restaurante : v) {
            if (restaurante.getPosicion().distancia(p) <= d) {
                if (mejor == null || restaurante.getValoracion() > mejor.getValoracion()) {
                    mejor = restaurante;
                }
            }
        }

        return mejor.getNombre();
    }

    public static void main(String[] args) {
        Restaurante[] restaurantes = {
            new Restaurante("Restaurante1", new Punto(10, 10), 4),
            new Restaurante("Restaurante2", new Punto(20, 20), 5),
            new Restaurante("Restaurante3", new Punto(30, 30), 3)
        };

        Punto usuario = new Punto(15, 15);
        double distanciaMax = 15;

        String mejorRestaurante = restauranteMejorValorado(restaurantes, usuario, distanciaMax);
        System.out.println("El mejor restaurante dentro de la distancia es: " + mejorRestaurante);
    }
}

