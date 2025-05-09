package Hector.Prac3;

public class Diccionario {
    private Nodo cabeza;
    public class Nodo{
        private String cadena;
        private int cantidad;
        Nodo sig;

        public Nodo(int cantidad, String cadena){
            this.cadena = cadena;
            this.cantidad = cantidad;
            this.sig = null;
        }

    }

    public Diccionario(){
        cabeza = null;
    }

    public void añadir(String unaCadena, int unaCantidad){
        if(cabeza == null || unaCadena.compareTo(cabeza.cadena) > 0){
            Nodo nuevo = new Nodo(unaCantidad, unaCadena);
            nuevo.sig = cabeza;
            cabeza = nuevo;
        }

        Nodo anterior = null;

        for(Nodo actual = cabeza; actual != null; actual = actual.sig){
            if(unaCadena.equals(actual.cadena)){
                actual.cantidad += unaCantidad;
                return;
            }

            if(unaCadena.compareTo(actual.cadena) < 0){
                Nodo nuevo = new Nodo(unaCantidad, unaCadena);
                nuevo.sig = actual;
                if(anterior != null){
                    anterior.sig = nuevo;
                }
                return;
            }
            anterior = actual;
        }
        Nodo nuevo = new Nodo(unaCantidad, unaCadena);
        anterior.sig = nuevo;
        
    }

    public String cadenaConMayorCantidad(){
        if(cabeza == null){
            return null;
        }
        Nodo actual = cabeza;
        int mayorCantidad = actual.cantidad;
        String mayorCadena = actual.cadena;

        for(; actual != null; actual = actual.sig){
            if(actual.cantidad > mayorCantidad){
                mayorCantidad = actual.cantidad;
                mayorCadena = actual.cadena;
            }
        }
        return mayorCadena;
    }
    public int cantidad(String unaCadena){
        Nodo actual = cabeza;
        for(; actual != null; actual = actual.sig){
            if(actual.cadena.equals(unaCadena)){
                return actual.cantidad;
            }
        }
        return 0;
    }

    public String toString(){
        String cadena ="{";
        for(Nodo actual = cabeza; actual != null; actual = actual.sig){
            cadena += actual.cadena + "---->" + actual.cantidad;
            if(actual.sig != null){
                cadena += ",";
            }
        }
        cadena += "}";
        return cadena;
    }


}
